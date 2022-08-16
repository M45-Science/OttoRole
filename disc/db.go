package disc

import (
	"RoleKeeper/cons"
	"RoleKeeper/rclog"
	"bytes"
	"compress/zlib"
	"encoding/json"
	"log"
	"os"
	"sync"
)

var (
	DBWriteLock sync.Mutex
	DBLock      sync.RWMutex
)

func compressZip(data []byte) []byte {
	var b bytes.Buffer
	w, err := zlib.NewWriterLevel(&b, zlib.BestSpeed)
	if err != nil {
		log.Println("ERROR: Gzip writer failure:", err)
	}
	w.Write(data)
	w.Close()
	return b.Bytes()
}

func DumpGuilds() {
	DBWriteLock.Lock()
	defer DBWriteLock.Unlock()

	fo, err := os.Create(cons.DBName)
	if err != nil {
		rclog.DoLog("Couldn't open db file, skipping...")
		return
	}
	/*  close fo on exit and check for its returned error */
	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()

	DBLock.RLock()

	rclog.DoLog("DumpGuilds: Writing guilds...")

	outbuf := new(bytes.Buffer)
	enc := json.NewEncoder(outbuf)
	if err := enc.Encode(Guilds); err != nil {
		rclog.DoLog("DumpGuilds: enc.Encode failure")
		return
	}
	DBLock.RUnlock()

	nfilename := cons.DBName + ".tmp"
	compBuf := compressZip(outbuf.Bytes())
	err = os.WriteFile(nfilename, compBuf, 0644)

	if err != nil {
		rclog.DoLog("DumpGuilds: Couldn't write db temp file.")
		return
	}

	oldName := nfilename
	newName := cons.DBName
	err = os.Rename(oldName, newName)

	if err != nil {
		rclog.DoLog("DumpGuilds: Couldn't rename db temp file.")
		return
	}

	rclog.DoLog("DumpGuilds: Complete!")
}
