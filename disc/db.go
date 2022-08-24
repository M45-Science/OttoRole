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
	"strconv"
	"sync"
	"time"
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

	startTime := time.Now()
	
	for i, c := range Clusters {
		if c == nil {
			os.Exit(1)
			return
		}
		WriteCluster(i)
	}
	endTime := time.Now()
	rclog.DoLog("DB Write Complete, took: " + endTime.Sub(startTime).String())
	os.Exit(1)
}

func WriteCluster(i int) {
	startTime := time.Now()

	cluster := Clusters[i]

	cluster.Lock.RLock()

	buf := new(bytes.Buffer)
	for _, g := range cluster.Guilds {

		if g == nil {
			break
		}
		binary.Write(buf, binary.LittleEndian, g.LID)
		binary.Write(buf, binary.LittleEndian, g.Customer)
		binary.Write(buf, binary.LittleEndian, g.Added)
		binary.Write(buf, binary.LittleEndian, g.Modified)
		binary.Write(buf, binary.LittleEndian, g.Donator)
		binary.Write(buf, binary.LittleEndian, g.Premium)
	}
	name := fmt.Sprintf("db/cluster-%v.dat", i+1)
	os.WriteFile(name, buf.Bytes(), 0644)

	defer cluster.Lock.RUnlock()
	endTime := time.Now()
	rclog.DoLog("Cluster-" + strconv.FormatInt(int64(i), 10) + " write, took: " + endTime.Sub(startTime).String())
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
