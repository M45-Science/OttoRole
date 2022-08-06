package glob

import (
	"os"
	"time"
)

var Uptime time.Time

var LogDesc *os.File
var LogName string

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

var GuildList map[uint64]GuildSettings
