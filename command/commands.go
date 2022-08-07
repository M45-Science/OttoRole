package command

import "github.com/bwmarrin/discordgo"

var cmds = []Command{
	{
		AppCmd: &discordgo.ApplicationCommand{
			Name:        "role",
			Description: "Assign roles to yourself!",
		},
	},
}
