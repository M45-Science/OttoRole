package logout

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"RoleKeeper/glob"
)

/* Normal CW log */
func DoLogMain(text string) {
	ctime := time.Now()
	_, filename, line, _ := runtime.Caller(1)

	date := fmt.Sprintf("%2v:%2v.%2v", ctime.Hour(), ctime.Minute(), ctime.Second())
	buf := fmt.Sprintf("%v: %15v:%5v: %v\n", date, filepath.Base(filename), line, text)
	_, err := glob.CWLogDesc.WriteString(buf)
	if err != nil {
		fmt.Println("DoLogMain: WriteString failure")
		glob.CWLogDesc.Close()
		glob.CWLogDesc = nil
		return
	}
}

/* Prep everything for the cw log */
func StartMainLog() {
	t := time.Now()

	/* Create our log file names */
	glob.CWLogName = fmt.Sprintf("log/cw-%v-%v-%v.log", t.Day(), t.Month(), t.Year())

	/* Make log directory */
	errr := os.MkdirAll("log", os.ModePerm)
	if errr != nil {
		fmt.Print(errr.Error())
		return
	}

	/* Open log files */
	bdesc, errb := os.OpenFile(glob.CWLogName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	/* Handle file errors */
	if errb != nil {
		fmt.Printf("An error occurred when attempting to create cw log. Details: %s", errb)
		return
	}

	/* Save descriptors, open/closed elsewhere */
	glob.CWLogDesc = bdesc
}
