package glob

import (
	"os"
	"time"
)

var (
	Uptime time.Time

	LogDesc *os.File
	LogName string

	ServerRunning        bool
	DoRegisterCommands   *bool
	DoDeregisterCommands *bool
	TestMode             *bool
)
