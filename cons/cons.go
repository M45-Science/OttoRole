package cons

import "time"

const (
	Version            = "002-010623-0103"
	BotName            = "RoleKeeper"
	DumpName           = "db-dump.json"
	ClusterPrefix      = "cluster-"
	ClusterSuffix      = ".ccf"
	ConfigFile         = "config.json"
	LIDTopFile         = "LIDTop.dat"
	MaxDiscordAttempts = 50
	NumClusters        = 128
	MaxGuilds          = 16384 //Preallocated for speed
	LimitRoles         = 8
	LockRest           = time.Millisecond

	/*GMT Thu Sep 01 2022 06:00:00 GMT+0000*/
	RoleKeeperEpoch = 1662012000
	/* 32bit records will roll over at */
	/* GMT: Wednesday, March 16, 2242 12:56:32 PM */
	/* So we need to remember to upgrade this in 220 years ;) */

	//Test only
	TSize = 134
)
