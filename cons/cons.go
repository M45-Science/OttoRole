package cons

const (
	Version            = "001-0904"
	BotName            = "RoleKeeper"
	DumpName           = "db-dump.json"
	ClusterPrefix      = "cluster-"
	ClusterSuffix      = ".ccf"
	ConfigFile         = "config.json"
	LIDTopFile         = "LIDTop.dat"
	MaxDiscordAttempts = 50
	NumClusters        = 128
	MaxGuilds          = 16384

	/*GMT Thu Sep 01 2022 06:00:00 GMT+0000*/
	RoleKeeperEpoch = 1662012000
	/* 32bit records will roll over at */
	/* GMT: Wednesday, March 16, 2242 12:56:32 PM */
	/* So we need to remember to upgrade this in 220 years ;) */

	RecordEnd = 0x21

	//Test only
	TSize = 134
)
