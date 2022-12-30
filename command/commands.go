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
	{
		AppCmd: &discordgo.ApplicationCommand{
			Name:        "add-role",
			Description: "Add or remove roles to the list",
		},
		Command: AddRole,
		ModOnly: true,
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
	var availableRoles []discordgo.SelectMenuOption
	roles := GetGuildRoles(s, i)
	for _, role := range roles {
		for _, existing := range guild.Roles {
			if db.IntToID(existing.ID) == role.ID {
				continue
			}
		}
		entry := discordgo.SelectMenuOption{Label: role.Name, Value: role.Name}
		availableRoles = append(availableRoles, entry)
	}
	if len(availableRoles) <= 0 {
		disc.EphemeralResponse(s, i, "Error:", "Sorry, there are no eligabile roles that can be added!")
		return
	}

	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Add a role that normal users should be allowed to self-assign:",
			Flags:   1 << 6,
			Components: []discordgo.MessageComponent{
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
		},
	}

	err := s.InteractionRespond(i.Interaction, response)
	if err != nil {
		cwlog.DoLog(err.Error())
	}
}

func GetGuildRoles(s *discordgo.Session, i *discordgo.InteractionCreate) []*discordgo.Role {
	guild, err := s.Guild(i.GuildID)
	if guild != nil && err == nil {
		return guild.Roles
	}
	return nil
}
