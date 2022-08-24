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
	//Name type bytes
	LID       uint32 //4
	Customer  uint64 //8
	Guild     uint64 //8
	Added     uint64 //8
	Modified  uint64 //8
	ReservedA uint64 //8

	Donator   uint16 //2
	Premium   uint16 //2
	ReservedB uint16 //2

	Roles []RoleData
	Lock  sync.RWMutex
}

// Total size + 2 for end
const RecordSize = 52

func IntToID(id uint64) string {
	strId := fmt.Sprintf("%v", id)
	return strId
}
