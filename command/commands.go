package command

import (
	"RoleKeeper/cwlog"
	"RoleKeeper/disc"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

var cmds = []Command{
	{
		AppCmd: &discordgo.ApplicationCommand{
			Name:        "roles",
			Description: "Add or Remove roles to yourself, for groups and notifcations!",
		},
		Command: RoleCommand,
	},
}

func RoleCommand(s *discordgo.Session, i *discordgo.InteractionCreate, guild *disc.GuildData) {
	buf := fmt.Sprintf("LID: %v", guild.LID)
	cwlog.DoLog(buf)
}
