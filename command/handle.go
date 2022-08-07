package command

import (
	"RoleKeeper/cfg"
	"RoleKeeper/rclog"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type Command struct {
	Command func(s *discordgo.Session, i *discordgo.InteractionCreate)
	AppCmd  *discordgo.ApplicationCommand

	AdminOnly bool
}

var CL []Command

func RegisterCommands(s *discordgo.Session, g string) {
	CL = cmds

	for i, o := range CL {
		if o.AppCmd == nil {
			continue
		}
		if o.AppCmd.Name == "" ||
			o.AppCmd.Description == "" {
			continue
		}

		cmd, err := s.ApplicationCommandCreate(cfg.Config.App, g, o.AppCmd)
		if err != nil {
			rclog.DoLog("Failed to create command: " +
				CL[i].AppCmd.Name)
			continue
		}
		CL[i].AppCmd = cmd
	}
}

func SlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.AppID != cfg.Config.App {
		return
	}

	if i.Member == nil {
		return
	}

	if i.Type == discordgo.InteractionApplicationCommand {
		data := i.ApplicationCommandData()

		if strings.EqualFold(data.Name, "role") {
			fmt.Println("Meep.")
		}
	}

}
