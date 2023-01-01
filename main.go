package main

import (
	"RoleKeeper/cfg"
	"RoleKeeper/command"
	"RoleKeeper/cons"
	"RoleKeeper/cwlog"
	"RoleKeeper/db"
	"RoleKeeper/disc"
	"RoleKeeper/glob"
	"flag"
	"fmt"
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

func main() {
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

	glob.ServerRunning = true
	glob.DoRegisterCommands = flag.Bool("regCommands", false, "Register discord commands")
	glob.DoDeregisterCommands = flag.Bool("deregCommands", false, "Deregister discord commands")
	glob.TestMode = flag.Bool("testMode", false, "WILL COMPLETELY OVER-WRITE CURRENT DB, AND GENERATE A FAKE TEST ONE.")
	flag.Parse()

	db.ThreadCount = runtime.NumCPU()
	debug.SetMemoryLimit(1024 * 1024 * 1024 * 24) //24gb
	debug.SetMaxThreads(db.ThreadCount * 2)

	glob.Uptime = time.Now().UTC().Round(time.Second)
	cwlog.StartLog()

	cfg.ReadCfg()
	cfg.WriteCfg()

	go startbot()

	/* Wait here for process signals */
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	tNow := time.Now().Round(time.Second)
	cwlog.DoLog("uptime: " + tNow.Sub(glob.Uptime).String())
	command.ClearCommands()
}

var DiscordConnectAttempts int

func startbot() {
	if cfg.Config.Token == "" {
		cwlog.DoLog("No discord token.")
		return
	}

	cwlog.DoLog(cons.BotName + " " + cons.Version + " starting.")
	cwlog.DoLog("Max Guilds: " + strconv.FormatInt(cons.MaxGuilds, 10))

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

	disc.Session = s
	disc.Ready = r
	cwlog.DoLog("Discord bot ready")

	db.GuildLookup = make(map[uint64]*db.GuildData, cons.TSize)

	if *glob.TestMode {
		testDatabase()
		db.WriteAllCluster()
		//disc.ReadAllClusters()
	} else {
		db.ReadAllClusters()
		//disc.WriteAllCluster()
	}
	db.UpdateGuildLookup()
	go db.LookupRoleNames(s, nil)

	if *glob.DoDeregisterCommands {
		command.RegisterCommands(s)
	}
	if *glob.DoDeregisterCommands {
		command.ClearCommands()
	}

	go MainLoop()

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

	cwlog.DoLog("Making test database...")

	tNow := db.NowToCompact()
	for x := 0; x < cons.TSize; x++ {

		//Make guild
		newGuild := db.GuildData{LID: uint32(x), Customer: rand.Uint64(), Guild: rand.Uint64(), Added: uint32(tNow), Modified: uint32(tNow), Donator: 0}
		db.Database[x] = &newGuild
	}
	db.LID_TOP = cons.TSize

	buf := fmt.Sprintf("Guilds: %v", db.LID_TOP)
	cwlog.DoLog(buf)
}
