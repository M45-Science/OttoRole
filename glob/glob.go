package glob

import (
	"os"
	"time"
)

var (
	Uptime time.Time

	LogDesc *os.File
	LogName string

	Guilds map[uint64]*GuildData
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
