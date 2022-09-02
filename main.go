package main

import (
	"RoleKeeper/cfg"
	"RoleKeeper/command"
	"RoleKeeper/cons"
	"RoleKeeper/cwlog"
	"RoleKeeper/disc"
	"RoleKeeper/glob"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"strconv"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

const version = "0.0.1"

func main() {

	glob.ServerRunning = true
	glob.DoRegisterCommands = flag.Bool("regCommands", false, "Register discord commands")
	glob.DoDeregisterCommands = flag.Bool("deregCommands", false, "Deregister discord commands")
	glob.LocalTestMode = flag.Bool("testMode", false, "WILL OVER-WRITE CURRENT DB, AND GENERATE A FAKE ONE.")
	flag.Parse()

	disc.ThreadCount = runtime.NumCPU()
	debug.SetMemoryLimit(1024 * 1024 * 1024 * 24)
	debug.SetMaxThreads(disc.ThreadCount * 2)

	glob.Uptime = time.Now().UTC().Round(time.Second)
	cwlog.StartLog()

	cfg.ReadCfg()
	cfg.WriteCfg()

	go startbot()

	/* Wait here for process signals */
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	command.ClearCommands()
}

var DiscordConnectAttempts int

func startbot() {
	if cfg.Config.Token == "" {
		cwlog.DoLog("No discord token.")
		return
	}

	cwlog.DoLog(cons.BotName + " " + version + " starting.")
	cwlog.DoLog("Max Guilds: " + strconv.FormatInt((cons.NumClusters*cons.ClusterSize), 10))

	bot, err := discordgo.New("Bot " + cfg.Config.Token)

	if err != nil {
		cwlog.DoLog(fmt.Sprintf("An error occurred when attempting to create the Discord session. Details: %v", err))
		time.Sleep(time.Minute * (5 * cons.MaxDiscordAttempts))
		DiscordConnectAttempts++

		if DiscordConnectAttempts < cons.MaxDiscordAttempts {
			startbot()
		}
		return
	}

	bot.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAllWithoutPrivileged)

	bot.AddHandler(botReady)
	errb := bot.Open()

	if errb != nil {
		cwlog.DoLog(fmt.Sprintf("An error occurred when attempting to create the Discord session. Details: %v", errb))
		time.Sleep(time.Minute * (5 * cons.MaxDiscordAttempts))
		DiscordConnectAttempts++

		if DiscordConnectAttempts < cons.MaxDiscordAttempts {
			startbot()
		}
		return
	}

	bot.LogLevel = discordgo.LogWarning
}

func botReady(s *discordgo.Session, r *discordgo.Ready) {

	botstatus := cfg.Config.Domain
	err := s.UpdateGameStatus(0, botstatus)
	if err != nil {
		cwlog.DoLog(err.Error())
	}

	s.AddHandler(command.SlashCommand)
	command.RegisterCommands(s)

	disc.Session = s
	disc.Ready = r
	cwlog.DoLog("Discord bot ready")

	disc.GuildLookup = make(map[uint64]*disc.GuildData, cons.TSize)

	cwlog.DoLog("Record Size: " + strconv.FormatInt(disc.RecordSize, 10) + "b")
	cwlog.DoLog("Cluster Size: " + strconv.FormatInt(disc.RecordSize*cons.ClusterSize+2, 10) + "b")

	if *glob.LocalTestMode {
		testDatabase()
		disc.WriteAllCluster()
		//disc.ReadAllClusters()
	} else {
		disc.ReadAllClusters()
		//disc.WriteAllCluster()
	}
	disc.UpdateGuildLookup()

	if *glob.DoDeregisterCommands {
		command.RegisterCommands(s)
	}

	disc.MainLoop()
}

func testDatabase() {
	os.RemoveAll("data/db")
	/* Make data directory */
	errr := os.MkdirAll("data", os.ModePerm)
	if errr != nil {
		fmt.Print(errr.Error())
		return
	}
	/* Make log directory */
	errr = os.MkdirAll("data/db", os.ModePerm)
	if errr != nil {
		fmt.Print(errr.Error())
		return
	}

	cwlog.DoLog("Making test map...")

	tNow := disc.NowToCompact()
	for x := 0; x < cons.TSize; x++ {

		//Make guild
		newGuild := disc.GuildData{LID: uint32(x), Customer: rand.Uint64(), Guild: rand.Uint64(), Added: uint32(tNow), Modified: uint32(tNow), Donator: 0}

		disc.Clusters[x%cons.NumClusters].Guilds[(x%cons.ClusterSize)/cons.NumClusters] = &newGuild
	}
	disc.LID_TOP = cons.TSize

	buf := fmt.Sprintf("Guilds: %v, Clusters: %v, ClusterSize: %v, MaxGuilds: %v",
		disc.LIDTop,
		int(math.Ceil(float64(cons.TSize)/float64(cons.ClusterSize))),
		cons.ClusterSize,
		cons.ClusterSize*cons.NumClusters)
	cwlog.DoLog(buf)
}
