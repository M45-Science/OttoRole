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
			Name:        "role",
			Description: "Assign roles to yourself!",
		},
		Command: RoleCommand,
	},
}

func RoleCommand(s *discordgo.Session, i *discordgo.InteractionCreate, guild *disc.GuildData) {
	buf := fmt.Sprintf("LID: %v", guild.LID)
	cwlog.DoLog(buf)
}
