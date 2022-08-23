package disc

import (
	"RoleKeeper/cons"
	"fmt"
	"sync"

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
	Lock   sync.RWMutex
}

type RoleData struct {
	Name string
	ID   uint64
}

type GuildData struct {
	LID      uint32
	Customer uint64
	Added    uint64
	Modified uint64
	Donator  uint8
	Premium  uint8
	Roles    []RoleData
}

func IntToID(id uint64) string {
	strId := fmt.Sprintf("%v", id)
	return strId
}
