package disc

import (
	"RoleKeeper/cons"
	"RoleKeeper/rclog"
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"encoding/json"
	"fmt"
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

func WriteAllCluster() {

	for i, c := range Clusters {
		if c == nil {
			os.Exit(1)
			return
		}
		WriteCluster(i)
	}
	os.Exit(1)
}

func WriteCluster(i int) {
	cluster := Clusters[i]

	if cluster == nil {
		return
	}

	cluster.Lock.RLock()

	for _, g := range cluster.Guilds {

		if g == nil {
			return
		}
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.LittleEndian, g.LID)
		binary.Write(buf, binary.LittleEndian, g.Customer)
		binary.Write(buf, binary.LittleEndian, g.Added)
		binary.Write(buf, binary.LittleEndian, g.Modified)
		binary.Write(buf, binary.LittleEndian, g.Donator)
		binary.Write(buf, binary.LittleEndian, g.Premium)
		fmt.Print(buf)
	}

	defer cluster.Lock.RUnlock()
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
	if err := enc.Encode(GuildLookup); err != nil {
		rclog.DoLog("DumpGuilds: enc.Encode failure")
		return
	}
	DBLock.RUnlock()

	nfilename := cons.DBName + ".tmp"
	//compBuf := compressZip(outbuf.Bytes())
	err = os.WriteFile(nfilename, outbuf.Bytes(), 0644)

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
