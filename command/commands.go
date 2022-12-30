package command

import (
	"RoleKeeper/cwlog"
	"RoleKeeper/db"
	"RoleKeeper/disc"

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

func RoleCommand(s *discordgo.Session, i *discordgo.InteractionCreate, guild *db.GuildData) {
	if len(guild.Roles) <= 0 {
		disc.EphemeralResponse(s, i, "ERROR:", "Sorry, there aren't any roles set up for this Discord guild right now!")
		return
	}
	for _, role := range guild.Roles {
		cwlog.DoLog(role.Name + "\n")
	}
}

func AddRole(s *discordgo.Session, i *discordgo.InteractionCreate, guid *db.GuildData) {
	//Send list of roles
}
