package command

import (
	"RoleKeeper/cwlog"
	"RoleKeeper/db"
	"RoleKeeper/disc"

	"github.com/bwmarrin/discordgo"
)

var modOnly int64 = discordgo.PermissionManageRoles

var cmds = []Command{
	{
		AppCmd: &discordgo.ApplicationCommand{
			Name:        "roles",
			Description: "Add or Remove roles to yourself, for groups and notifcations!",
		},
		Command: RoleCommand,
	},
	{
		AppCmd: &discordgo.ApplicationCommand{
			Name:                     "Add-role",
			Description:              "Add or remove roles to the list",
			DefaultMemberPermissions: &modOnly,
		},
		Command: AddRole,
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

func AddRole(s *discordgo.Session, i *discordgo.InteractionCreate, guild *db.GuildData) {
}
