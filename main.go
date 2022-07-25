package main

import (
	"RoleKeeper/glob"
	"time"
)

const version = "0.0.001"

func main() {

	glob.Uptime = time.Now()
}
