package cwlog

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"RoleKeeper/glob"
)

func DoLog(text string) {
	ctime := time.Now()
	_, filename, line, _ := runtime.Caller(1)

	date := fmt.Sprintf("%2v:%2v.%2v", ctime.Hour(), ctime.Minute(), ctime.Second())
	buf := fmt.Sprintf("%v: %15v:%5v: %v\n", date, filepath.Base(filename), line, text)
	_, err := glob.LogDesc.WriteString(buf)
	fmt.Print(buf)

	if err != nil {
		fmt.Println("DoLog: WriteString failure")
		glob.LogDesc.Close()
		glob.LogDesc = nil
		return
	}
}

/* Prep everything for the cw log */
func StartLog() {
	t := time.Now()

	/* Create our log file names */
	glob.LogName = fmt.Sprintf("logs/cw-%v-%v-%v.log", t.Day(), t.Month(), t.Year())

	/* Make log directory */
	errr := os.MkdirAll("logs", os.ModePerm)
	if errr != nil {
		fmt.Print(errr.Error())
		return
	}

	/* Open log files */
	bdesc, errb := os.OpenFile(glob.LogName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	/* Handle file errors */
	if errb != nil {
		fmt.Printf("An error occurred when attempting to create the log. Details: %s", errb)
		return
	}

	/* Save descriptors, used/closed elsewhere */
	glob.LogDesc = bdesc
}
