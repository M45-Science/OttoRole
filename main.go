package main

import (
	"RoleKeeper/cfg"
	"RoleKeeper/command"
	"RoleKeeper/cons"
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

	bot.AddHandler(BotReady)
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

func BotReady(s *discordgo.Session, r *discordgo.Ready) {
	botstatus := "https://" + cfg.Config.Domain
	err := s.UpdateGameStatus(0, botstatus)
	if err != nil {
		rclog.DoLog(err.Error())
	}

	s.AddHandler(command.SlashCommand)
	command.RegisterCommands(s, "916844097883471923")

	rclog.DoLog("Discord bot ready")
	rclog.DoLog("Making test map...")

	var tSize uint64 = 1000000
	var i uint64
	glob.Guilds = make(map[uint64]*glob.GuildData, tSize)
	tnow := time.Now().Unix()
	tRoles := []glob.RoleData{}
	//Make some role data
	for i = 0; i < 15; i++ {
		rid := rand.Uint64()
		tRoles = append(tRoles, glob.RoleData{Name: fmt.Sprintf("%012d", rid), ID: rid})
	}

	start := time.Now()

	var rid uint64
	var idlist []uint64
	//Test map
	for i = 0; i < tSize; i++ {
		rid = rand.Uint64()
		if glob.Guilds[rid] == nil {
			newGuild := glob.GuildData{Added: tnow, Modified: tnow, Donator: 0, Premium: 0, Roles: tRoles}
			glob.Guilds[rid] = &newGuild
		}
		if i%100 == 0 {
			idlist = append(idlist, rid)
		}
	}

	end := time.Now()

	rclog.DoLog("Make map took: " + end.Sub(start).String())

	start = time.Now()
	var GetData uint64
	for _, i := range idlist {
		GetData = glob.Guilds[i].Roles[1].ID
	}
	end = time.Now()

	if GetData != 0 {
		//
	}
	rclog.DoLog("Lookup took: " + end.Sub(start).String())
	rclog.DoLog(fmt.Sprintf("Lookups: %v", len(idlist)))
}
