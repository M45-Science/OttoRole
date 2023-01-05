package command

import (
	"RoleKeeper/cfg"
	"RoleKeeper/cwlog"
	"RoleKeeper/db"
	"RoleKeeper/disc"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func SlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {

	/* Ignore possible malicious or erroneous */
	if i.AppID != cfg.Config.App {
		return
	}

	/* Ignore DMs */
	if i.Member == nil {
		return
	}

	/* Handle componet-responses */
	if i.Type == discordgo.InteractionMessageComponent {
		handleComponet(s, i)
		return
	} else if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	/*
	 * Lookup guild, add to DB if not found
	 */
	g := db.GuildLookupReadString(i.GuildID)
	if g == nil {
		/* Add to db */
		gid, err := db.SnowflakeToInt(i.GuildID)
		if err == nil {
			db.AddGuild(gid)
		} else {
			cwlog.DoLog(fmt.Sprintf("Failed to parse guildid: %v", i.GuildID))
			return
		}
	}
	g = db.GuildLookupReadString(i.GuildID)

	data := i.ApplicationCommandData()
	CmdName := data.Name

	/* Run standard slash commands */
	for _, c := range cmds {
		if c.AppCmd.Name == CmdName {
			if c.AdminOnly && i.Member.Permissions < discordgo.PermissionAdministrator {
				disc.EphemeralResponse(s, i, disc.DiscRed, "ERROR:", "You do not have the necessary permissions to use this command.", false)
				return
			} else if c.ModOnly && i.Member.Permissions < discordgo.PermissionManageRoles {
				disc.EphemeralResponse(s, i, disc.DiscRed, "ERROR:", "You do not have the necessary permissions to use this command.", false)
				return
			}
			c.Command(s, i, g)
			return
		}
	}

}
