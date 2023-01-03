package command

import (
	"RoleKeeper/cwlog"
	"RoleKeeper/db"
	"RoleKeeper/disc"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func handleComponet(s *discordgo.Session, i *discordgo.InteractionCreate) {

	data := i.MessageComponentData()

	if strings.EqualFold(data.CustomID, "config-roles") {

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

			disc.EphemeralResponse(s, i, disc.DiscPurple, "Status:", "Working...")

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
}
