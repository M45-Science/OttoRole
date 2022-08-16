package disc

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

var (
	Guilds  map[uint64]*GuildData
	Session *discordgo.Session
	Ready   *discordgo.Ready
)

type RoleData struct {
	Name string
	ID   uint64
}

type GuildData struct {
	Added    int64
	Modified int64
	Donator  uint8
	Premium  uint8
	Roles    []RoleData
}

func IntToID(id uint64) string {
	strId := fmt.Sprintf("%v", id)
	return strId
}

