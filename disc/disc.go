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
	LID      uint32 `json:"L"` //4
	Customer uint64 `json:"C"` //8
	Guild    uint64 `json:"G"` //8
	Added    uint32 `json:"A"` //4
	Modified uint32 `json:"M"` //4

	Donator uint8 `json:"D"` //1

	Roles []RoleData   `json:"-"`
	Lock  sync.RWMutex `json:"-"`
}

func IntToID(id uint64) string {
	strId := fmt.Sprintf("%v", id)
	return strId
}
