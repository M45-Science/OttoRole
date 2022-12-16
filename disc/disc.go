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
	Database [cons.MaxGuilds]*GuildData
)

type RoleData struct {
	Name string
	ID   uint64
}

type GuildData struct {
	//Name type bytes
	LID      uint32 `json:"l,omitempty"` //4
	Customer uint64 `json:"c,omitempty"` //8
	Guild    uint64 `json:"-"`           //8 --Already in JSON as KEY
	Added    uint32 `json:"a,omitempty"` //4
	Modified uint32 `json:"m,omitempty"` //4

	Donator uint8 `json:"d,omitempty"` //1

	/* Not on disk */
	Roles []RoleData   `json:"-"`
	Lock  sync.RWMutex `json:"-"`
}

func IntToID(id uint64) string {
	strId := fmt.Sprintf("%v", id)
	return strId
}
