package command

import (
	"RoleKeeper/cfg"
	"RoleKeeper/cwlog"
	"RoleKeeper/db"
	"RoleKeeper/disc"
	"RoleKeeper/glob"
	"fmt"

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

	/* App commands only */
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	g := db.GuildLookupReadString(i.GuildID)
	/* Ignore guilds not in our DB */
	if g == nil {
		EphemeralResponse(s, i, 0xFF0000, "Error:", "Sorry, this guild isn't registered yet!")

		buf := fmt.Sprintf("Guild not found: %v", i.GuildID)
		cwlog.DoLog(buf)

		/* Add to db */
		gid, err := db.GuildStrToInt(i.GuildID)
		if err == nil {
			db.AddGuild(gid)
		} else {
			cwlog.DoLog(fmt.Sprintf("Failed to parse guildid: %v", i.GuildID))
		}
		return
	}

	data := i.ApplicationCommandData()
	CmdName := data.Name

	/* Ignore empty command IDs */
	if CmdName == "" {
		return
	}

	for _, c := range cmds {
		if c.AppCmd.Name == CmdName {
			if c.AdminOnly && i.Member.Permissions < discordgo.PermissionAdministrator {
				EphemeralResponse(s, i, 1, "ERROR:", "You do not have the proper permissions to use this command.")
				return
			} else if c.ModOnly && i.Member.Permissions < discordgo.PermissionManageRoles {
				EphemeralResponse(s, i, 1, "ERROR:", "You do not have the proper permissions to use this command.")
				return
			}
			c.Command(s, i, g)
			return
		}
	}

}

func EphemeralResponse(s *discordgo.Session, i *discordgo.InteractionCreate, color int, title, message string) {
	//cwlog.DoLog("EphemeralResponse:\n" + i.Member.User.Username + "\n" + title + "\n" + message)

	var elist []*discordgo.MessageEmbed
	elist = append(elist, &discordgo.MessageEmbed{Title: title, Description: message, Color: color})

	//1 << 6 is ephemeral/private
	respData := &discordgo.InteractionResponseData{Embeds: elist, Flags: 1 << 6}
	resp := &discordgo.InteractionResponse{Type: discordgo.InteractionResponseChannelMessageWithSource, Data: respData}
	err := s.InteractionRespond(i.Interaction, resp)
	if err != nil {
		cwlog.DoLog(err.Error())
	}
}
