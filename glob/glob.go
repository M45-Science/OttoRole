package glob

import (
	"os"
	"time"
)

var (
	Uptime time.Time

	LogDesc *os.File
	LogName string

	GuildList map[uint64]GuildSettings
)

type roleData struct {
	Name string
	ID   uint64
}

type GuildSettings struct {
	Added    int64
	Modified int64
	Premium  int
	Roles    []roleData
}
