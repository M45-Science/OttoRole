package main

import (
	"RoleKeeper/cfg"
	"RoleKeeper/command"
	"RoleKeeper/cons"
	"RoleKeeper/disc"
	"RoleKeeper/glob"
	"RoleKeeper/rclog"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

const version = "0.0.1"

func main() {

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

	//time.Sleep(5 * time.Second)
	testDatabase()
	//disc.DumpGuilds()
	disc.WriteAllCluster()
}

func testDatabase() {
	rclog.DoLog("Making test map...")

	var tSize uint64 = 100000000
	var x uint64
	//var y uint64
	disc.GuildLookup = make(map[uint64]*disc.GuildData, tSize)
	tnow := time.Now().Unix()
	var rid uint64

	//Test map
	for x = 0; x < tSize; x++ {
		rid = rand.Uint64()
		if disc.GuildLookup[rid] == nil {

			tRoles := []disc.RoleData{}

			//Make some role data
			/* for y = 1; y <= 5; y++ {
				rid := rand.Uint64()
				tRoles = append(tRoles, disc.RoleData{Name: "role" + disc.IntToID(y), ID: rid})
			} */

			newGuild := disc.GuildData{LID: uint32(x), Customer: rand.Uint64(), Added: uint64(tnow), Modified: uint64(tnow), Donator: 0, Premium: 0, Roles: tRoles}
			disc.GuildLookup[rid] = &newGuild

			if x%cons.ClusterSize == 0 {
				if disc.Clusters[disc.ClusterTop/cons.ClusterSize] == nil {
					disc.Clusters[disc.ClusterTop/cons.ClusterSize] =
						&disc.ClusterData{}
					buf := fmt.Sprintf("New Cluster: %v", (disc.ClusterTop/cons.ClusterSize)+1)
					rclog.DoLog(buf)
				}
			}
			disc.Clusters[disc.ClusterTop/cons.ClusterSize].Guilds[disc.ClusterTop%cons.ClusterSize] = &newGuild
			disc.ClusterTop++
		}
	}
	buf := fmt.Sprintf("KGuilds: %v, Clusters: %v, ClusterSize: %v, Max-MGuilds: %0.2f",
		disc.ClusterTop/1000,
		(disc.ClusterTop / cons.ClusterSize),
		cons.ClusterSize,
		cons.ClusterSize*cons.MaxClusters/1000000.0)
	rclog.DoLog(buf)
}
