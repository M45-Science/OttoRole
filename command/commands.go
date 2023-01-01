package command

import (
	"RoleKeeper/db"
	"RoleKeeper/disc"
	"strings"

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
	{
		AppCmd: &discordgo.ApplicationCommand{
			Name:        "config-roles",
			Description: "Add or remove roles to the list",
		},
		Command: AddRole,
		ModOnly: true,
	},
}

func RoleCommand(s *discordgo.Session, i *discordgo.InteractionCreate, guild *db.GuildData) {
	if len(guild.Roles) == 0 {
		disc.EphemeralResponse(s, i, disc.DiscOrange, "ERROR:", "Sorry, there aren't any roles set up for this Discord guild right now!")
		return
	}
	buf := ""
	for c, role := range guild.Roles {
		if c > 0 {
			buf = buf + ", "
		}
		if role.Name == "" {
			buf = buf + db.IntToSnowflake(role.ID)
		}
		buf = buf + role.Name
	}
	disc.EphemeralResponse(s, i, disc.DiscOrange, "Test:", "```"+buf+"```")
}

func AddRole(s *discordgo.Session, i *discordgo.InteractionCreate, guild *db.GuildData) {

	disc.EphemeralResponse(s, i, disc.DiscPurple, "Status:", "Finding eligible roles.")

	var availableRoles []discordgo.SelectMenuOption
	roles := disc.GetGuildRoles(s, i.GuildID)

	for _, role := range roles {

		//Block specific names
		if strings.EqualFold(role.Name, "@everyone") {
			continue
		}

		found := false

		//Exclude roles with moderator permissions
		if role.Permissions&(discordgo.PermissionAdministrator|
			discordgo.PermissionBanMembers|
			discordgo.PermissionManageRoles|
			discordgo.PermissionModerateMembers|
			discordgo.PermissionManageWebhooks|
			discordgo.PermissionManageServer) != 0 {
			found = true
		}

		//Show roles already in database
		for _, existing := range guild.Roles {
			if db.IntToSnowflake(existing.ID) == role.ID {

				entry := discordgo.SelectMenuOption{
					Emoji: discordgo.ComponentEmoji{
						Name: "✅",
					},
					Label: role.Name, Value: role.ID}
				availableRoles = append(availableRoles, entry)
				found = true
				break
			}
		}

		if !found {
			entry := discordgo.SelectMenuOption{Label: role.Name, Value: role.ID}
			availableRoles = append(availableRoles, entry)
		}

	}

	if len(availableRoles) <= 0 {
		disc.EphemeralResponse(s, i, disc.DiscRed, "Error:", "Sorry, there are no eligabile roles that can be added!")
		return
	}

	embed := []*discordgo.MessageEmbed{{
		Title:       "Info:",
		Description: "Select roles to add them to the list.\nRoles in the list can be self-assigned by users.\n\n✅ = Already in list, selecting will remove.\nRoles with moderator permissions are excluded.",
	}}
	respose := &discordgo.WebhookEdit{
		Components: &[]discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.SelectMenu{
						// Select menu, as other components, must have a customID, so we set it to this value.
						CustomID:    "AddRole",
						Placeholder: "Select one",
						Options:     availableRoles,
					},
				},
			},
		},
		Embeds: &embed,
	}

	s.InteractionResponseEdit(i.Interaction, respose)
}
