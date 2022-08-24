package main

import (
	"RoleKeeper/cfg"
	"RoleKeeper/command"
	"RoleKeeper/cons"
	"RoleKeeper/disc"
	"RoleKeeper/glob"
	"RoleKeeper/rclog"
	"fmt"
	"io/fs"
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

	disc.ThreadCount = runtime.NumCPU()
	debug.SetMemoryLimit(1024 * 1024 * 1024 * 24)
	debug.SetMaxThreads(disc.ThreadCount * 4)

	glob.Uptime = time.Now().UTC().Round(time.Second)
	rclog.StartLog()

	cfg.ReadCfg()
	cfg.WriteCfg()

	go startbot()

	/* Wait here for process signals */
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

}

var DiscordConnectAttempts int

func startbot() {
	if cfg.Config.Token == "" {
		rclog.DoLog("No discord token.")
		return
	}

	rclog.DoLog("RoleKeeper " + version + " starting.")
	rclog.DoLog("Max Guilds: " + strconv.FormatInt((cons.MaxClusters*cons.ClusterSize), 10))
	time.Sleep(3)

	bot, err := discordgo.New("Bot " + cfg.Config.Token)

	if err != nil {
		rclog.DoLog(fmt.Sprintf("An error occurred when attempting to create the Discord session. Details: %v", err))
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
		rclog.DoLog(fmt.Sprintf("An error occurred when attempting to create the Discord session. Details: %v", errb))
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

	botstatus := "https://" + cfg.Config.Domain
	err := s.UpdateGameStatus(0, botstatus)
	if err != nil {
		rclog.DoLog(err.Error())
	}

	s.AddHandler(command.SlashCommand)
	command.RegisterCommands(s, cfg.Config.App)

	disc.Session = s
	disc.Ready = r
	rclog.DoLog("Discord bot ready")

	disc.GuildLookup = make(map[uint64]*disc.GuildData, cons.TSize)

	rclog.DoLog("Record Size: " + strconv.FormatInt(disc.RecordSize, 10) + "b")
	rclog.DoLog("Cluster Size: " + strconv.FormatInt(disc.RecordSize*cons.ClusterSize+2, 10) + "b")

	if 1 == 1 {
		testDatabase()
		disc.WriteAllCluster()
		disc.ReadAllClusters()
	} else {
		disc.ReadAllClusters()
		disc.WriteAllCluster()
	}
	disc.UpdateGuildLookup()
}

func testDatabase() {
	os.RemoveAll("db")
	os.Mkdir("db", fs.ModePerm)
	rclog.DoLog("Making test map...")

	var x int
	var y int

	//Test map
	for x = 0; x < int(math.Ceil(float64(cons.TSize)/float64(cons.ClusterSize))); x++ {

		disc.Clusters[x] =
			&disc.ClusterData{}
		//buf := fmt.Sprintf("New Cluster: %v", (x)+1)
		//rclog.DoLog(buf)

		tnow := time.Now().Unix()
		for y = 0; y < cons.ClusterSize; y++ {
			/*
				tRoles := []disc.RoleData{}
				for y = 1; y <= 5; y++ {
					rid := rand.Uint64()
					tRoles = append(tRoles, disc.RoleData{Name: "role" + disc.IntToID(y), ID: rid})
				}
			*/

			newGuild := disc.GuildData{LID: uint32((x * cons.ClusterSize) + y), Customer: rand.Uint64(), Guild: rand.Uint64(), Added: uint64(tnow), Modified: uint64(tnow), Donator: 0, Premium: 0}

			disc.Clusters[x].Guilds[y] = &newGuild
			disc.ClusterTop++
			if disc.ClusterTop > cons.TSize {
				break
			}
		}
	}

	buf := fmt.Sprintf("Guilds: %v, Clusters: %v, ClusterSize: %v, Max-MGuilds: %0.2f",
		disc.ClusterTop,
		int(math.Ceil(float64(cons.TSize)/float64(cons.ClusterSize))),
		cons.ClusterSize,
		cons.ClusterSize*cons.MaxClusters/1000000.0)
	rclog.DoLog(buf)
}
