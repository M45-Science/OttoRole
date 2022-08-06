package main

import (
	"RoleKeeper/cfg"
	"RoleKeeper/glob"
	"RoleKeeper/rclog"
	"time"
)

const version = "0.0.1"

func main() {

	glob.Uptime = time.Now().UTC().Round(time.Second)
	rclog.StartLog()
	rclog.DoLog("RoleKeeper " + version + " starting.")

	cfg.ReadCfg()
	cfg.WriteCfg()

	go startbot()

}

func startbot() {
	if cfg.Config.Token == "" {
		rclog.DoLog("No discord token.")
		return
	}
}
