package command

import "github.com/bwmarrin/discordgo"

var cmds = []Command{
	{
		AppCmd: &discordgo.ApplicationCommand{
			Name:        "role",
			Description: "Assign roles to yourself!",
		},
		Command: RoleCommand,
	},
}

func RoleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	
}
