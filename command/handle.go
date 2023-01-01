package command

import (
	"RoleKeeper/cfg"
	"RoleKeeper/cwlog"
	"RoleKeeper/db"
	"RoleKeeper/disc"
	"RoleKeeper/glob"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type Command struct {
	Command func(s *discordgo.Session, i *discordgo.InteractionCreate, guild *db.GuildData)
	AppCmd  *discordgo.ApplicationCommand

	AdminOnly bool
	ModOnly   bool
}

var CL []Command

var adminPerms int64 = discordgo.PermissionAdministrator
var modPerms int64 = discordgo.PermissionManageRoles
var defaultPerms int64 = discordgo.PermissionUseSlashCommands

func RegisterCommands(s *discordgo.Session) {
	CL = cmds

	for i, o := range CL {

		if o.AdminOnly {
			o.AppCmd.DefaultMemberPermissions = &adminPerms
		} else if o.ModOnly {
			o.AppCmd.DefaultMemberPermissions = &modPerms
		} else {
			o.AppCmd.DefaultMemberPermissions = &defaultPerms
		}

		cmd, err := s.ApplicationCommandCreate(cfg.Config.App, "", o.AppCmd)
		if err != nil {
			cwlog.DoLog("Failed to create command: " + CL[i].AppCmd.Name)
			continue
		} else {
			cwlog.DoLog("Registered command: " + CL[i].AppCmd.Name)
		}
		CL[i].AppCmd = cmd
	}
}

func ClearCommands() {
	if *glob.DoDeregisterCommands && disc.Session != nil {
		cmds, _ := disc.Session.ApplicationCommands(cfg.Config.App, "")
		for _, v := range cmds {
			cwlog.DoLog(fmt.Sprintf("Deregistered command: %s", v.Name))
			err := disc.Session.ApplicationCommandDelete(disc.Session.State.User.ID, "", v.ID)
			if err != nil {
				cwlog.DoLog(err.Error())
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

	if i.Type == discordgo.InteractionMessageComponent {
		data := i.MessageComponentData()

		for _, c := range data.Values {
			if strings.EqualFold(data.CustomID, "AddRole") {
				//TODO: Check IDs and permissions

				guild := db.GuildLookupReadString(i.GuildID)
				if guild == nil {
					disc.EphemeralResponse(s, i, disc.DiscRed, "ERROR:", "This Discord guild is not in our database.")
					return
				}

				roleid, err := db.SnowflakeToInt(c)

				if err != nil {
					disc.EphemeralResponse(s, i, disc.DiscRed, "ERROR:", "Internal Error: The menu selection had invalid data.")
					return
				}

				found := false
				for _, role := range guild.Roles {
					if role.ID == roleid {
						//Already in DB, ignore
						found = true
						break
					}
				}

				if found {
					disc.EphemeralResponse(s, i, disc.DiscRed, "ERROR:", "That role is already in the list.")
					return
				}

				nguild, err := s.Guild(i.GuildID)
				if err != nil {
					disc.EphemeralResponse(s, i, disc.DiscRed, "ERROR:", "Internal Error: Unable to lookup this Discord guild.")
					break
				}

				numRoles := len(guild.Roles)
				if numRoles < 0xFF {

					for _, roleLookup := range nguild.Roles {
						if roleLookup.ID == db.IntToID(roleid) {
							newRole := db.RoleData{Name: roleLookup.Name, ID: roleid}

							guild.Lock.Lock()
							guild.Roles = append(guild.Roles, newRole)
							guild.Modified = db.NowToCompact()
							guild.Lock.Unlock()
							db.WriteAllCluster()

							disc.EphemeralResponse(s, i, disc.DiscGreen, "Status:", "Role added.")
							break
						}
					}
				}

			}
		}
		return
	} else if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

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

	/* Ignore empty command IDs */
	if CmdName == "" {
		return
	}

	for _, c := range cmds {
		if c.AppCmd.Name == CmdName {
			if c.AdminOnly && i.Member.Permissions < discordgo.PermissionAdministrator {
				disc.EphemeralResponse(s, i, disc.DiscRed, "ERROR:", "You do not have the necessary permissions to use this command.")
				return
			} else if c.ModOnly && i.Member.Permissions < discordgo.PermissionManageRoles {
				disc.EphemeralResponse(s, i, disc.DiscRed, "ERROR:", "You do not have the necessary permissions to use this command.")
				return
			}
			c.Command(s, i, g)
			return
		}
	}

}
