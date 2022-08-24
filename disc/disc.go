package disc

import (
	"RoleKeeper/cons"
	"fmt"
	"sync"

	"github.com/bwmarrin/discordgo"
)

var (
	ThreadCount     int
	GuildLookup     map[uint64]*GuildData
	GuildLookupLock sync.RWMutex

	Session    *discordgo.Session
	Ready      *discordgo.Ready
	Clusters   [cons.MaxClusters]*ClusterData
	ClusterTop int
)

type ClusterData struct {
	Guilds [cons.ClusterSize]*GuildData
}

type RoleData struct {
	Name string
	ID   uint64
}

type GuildData struct {
	LID      uint32
	Customer uint64
	Guild    uint64
	Added    uint64
	Modified uint64
	Donator  uint8
	Premium  uint8
	Roles    []RoleData
	Lock     sync.RWMutex
}

func IntToID(id uint64) string {
	strId := fmt.Sprintf("%v", id)
	return strId
}
