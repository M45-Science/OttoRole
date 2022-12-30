package disc

import (
	"RoleKeeper/cwlog"

	"github.com/bwmarrin/discordgo"
)

var (
	Session *discordgo.Session
	Ready   *discordgo.Ready
)

func InteractionResponse(s *discordgo.Session, i *discordgo.InteractionCreate, embed *discordgo.MessageEmbed) {
	cwlog.DoLog("InteractionResponse:\n" + i.Member.User.Username + "\n" + embed.Title + "\n" + embed.Description)

	var embedList []*discordgo.MessageEmbed
	embedList = append(embedList, embed)
	respData := &discordgo.InteractionResponseData{Embeds: embedList}
	resp := &discordgo.InteractionResponse{Type: discordgo.InteractionResponseChannelMessageWithSource, Data: respData}
	err := s.InteractionRespond(i.Interaction, resp)
	if err != nil {
		cwlog.DoLog(err.Error())
	}
}

func FollowupResponse(s *discordgo.Session, i *discordgo.InteractionCreate, f *discordgo.WebhookParams) {
	if f.Embeds != nil {
		cwlog.DoLog("FollowupResponse:\n" + i.Member.User.Username + "\n" + f.Embeds[0].Title + "\n" + f.Embeds[0].Description)

		_, err := s.FollowupMessageCreate(i.Interaction, false, f)
		if err != nil {
			cwlog.DoLog(err.Error())
		}
	} else if f.Content != "" {
		cwlog.DoLog("FollowupResponse:\n" + i.Member.User.Username + "\n" + f.Content)

		_, err := s.FollowupMessageCreate(i.Interaction, false, f)
		if err != nil {
			cwlog.DoLog(err.Error())
		}
	}

}

func EphemeralResponse(s *discordgo.Session, i *discordgo.InteractionCreate, title, message string) {
	cwlog.DoLog("EphemeralResponse:\n" + i.Member.User.Username + "\n" + title + "\n" + message)

	var elist []*discordgo.MessageEmbed
	elist = append(elist, &discordgo.MessageEmbed{Title: title, Description: message})

	//1 << 6 is ephemeral/private
	respData := &discordgo.InteractionResponseData{Embeds: elist, Flags: 1 << 6}
	resp := &discordgo.InteractionResponse{Type: discordgo.InteractionResponseChannelMessageWithSource, Data: respData}
	err := s.InteractionRespond(i.Interaction, resp)
	if err != nil {
		cwlog.DoLog(err.Error())
	}
}
