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

		if strings.EqualFold(data.CustomID, "AddRole") {

			for _, c := range data.Values {
				//TODO: Check IDs and permissions
				//Also re-check role ermissions
				if strings.HasSuffix(c, "-ignore") {
					disc.EphemeralResponse(s, i, disc.DiscRed, "ERROR:", "This role has moderatator permissions.\nLetting normal users self-assign a moderator role would be unadvisable.")
					return
				}

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

				disc.EphemeralResponse(s, i, disc.DiscGreen, "Status:", "Working...")

				found := -1
				for rpos, role := range guild.Roles {
					if role.ID == roleid {
						found = rpos
						break
					}
				}

				/*
				 * If role exists, remove it
				 */
				if found != -1 {
					guild.Lock.Lock()
					roleName := guild.Roles[found].Name
					if roleName == "" {
						roleName = db.IntToSnowflake(guild.Roles[found].ID)
					}
					guild.Roles = append(guild.Roles[:found], guild.Roles[found+1:]...)
					guild.Lock.Unlock()
					db.LookupRoleNames(s, guild)

					embed := []*discordgo.MessageEmbed{{
						Title:       "Status:",
						Description: "Role removed: " + roleName,
						Color:       disc.DiscOrange,
					}}
					respose := &discordgo.WebhookEdit{
						Embeds: &embed,
					}
					_, err = s.InteractionResponseEdit(i.Interaction, respose)
					if err != nil {
						cwlog.DoLog("Error: " + err.Error())
					}

					db.WriteAllCluster()
					db.DumpGuilds()
					return
				}

				/* Add role, if list isn't full */
				numRoles := len(guild.Roles)
				if numRoles < db.MAXROLES {
					newRole := db.RoleData{ID: roleid}
					guild.Lock.Lock()
					guild.Roles = append(guild.Roles, newRole)
					guild.Modified = db.NowToCompact()
					guild.Lock.Unlock()

					db.LookupRoleNames(s, guild)
					guild.Lock.RLock()
					roleName := db.IntToSnowflake(roleid)
					for _, role := range guild.Roles {
						if role.ID == roleid {
							roleName = role.Name
						}
					}
					guild.Lock.RUnlock()

					embed := []*discordgo.MessageEmbed{{
						Title:       "Status:",
						Description: "Role added: " + roleName,
						Color:       disc.DiscGreen,
					}}
					respose := &discordgo.WebhookEdit{
						Embeds: &embed,
					}
					_, err = s.InteractionResponseEdit(i.Interaction, respose)
					if err != nil {
						cwlog.DoLog("Error: " + err.Error())
					}

					db.WriteAllCluster()
					db.DumpGuilds()
					break
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
