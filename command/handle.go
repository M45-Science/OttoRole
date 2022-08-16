package command

import (
	"RoleKeeper/cfg"
	"RoleKeeper/disc"
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
		cmd, err := s.ApplicationCommandCreate(cfg.Config.App, "", o.AppCmd)
		if err != nil {
			rclog.DoLog("Failed to create command: " + CL[i].AppCmd.Name)
			continue
		}
		CL[i].AppCmd = cmd
	}
}

func ClearCommands() {
	if /**glob.DoDeregisterCommands && */ disc.Session != nil {
		cmds, _ := disc.Session.ApplicationCommands(cfg.Config.App, "")
		for _, v := range cmds {
			rclog.DoLog(fmt.Sprintf("Deregistered command: %s", v.Name))
			err := disc.Session.ApplicationCommandDelete(disc.Session.State.User.ID, "", v.ID)
			if err != nil {
				rclog.DoLog(err.Error())
			}
		}
	}
}

func SlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {

	/* Ignore possible malicious or erroneous */
	if i.AppID != cfg.Config.App {
		return
	}

	/* Ignore DMs */
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
