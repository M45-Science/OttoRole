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

	Session  *discordgo.Session
	Ready    *discordgo.Ready
	Clusters [cons.NumClusters]ClusterData
	LIDTop   int
)

type ClusterData struct {
	Guilds [cons.ClusterSize]*GuildData
}

type RoleData struct {
	Name string
	ID   uint64
}

type GuildData struct {
	//Name type bytes
	LID      uint32 //4
	Customer uint64 //8
	Guild    uint64 //8
	Added    uint32 //4
	Modified uint32 //4

	Donator uint8 //8

	Roles []RoleData
	Lock  sync.RWMutex
}

func IntToID(id uint64) string {
	strId := fmt.Sprintf("%v", id)
	return strId
}
