package disc

import (
	"RoleKeeper/cons"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

var (
	GuildLookup map[uint64]*GuildData
	Session     *discordgo.Session
	Ready       *discordgo.Ready
	Clusters    [cons.MaxClusters]*ClusterData
	ClusterTop  int
)

type ClusterData struct {
	Guilds [cons.ClusterSize]*GuildData
}

type RoleData struct {
	Name string
	ID   uint64
}

type GuildData struct {
	Customer uint64
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
