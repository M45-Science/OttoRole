package disc

import (
	"RoleKeeper/cons"
	"RoleKeeper/cwlog"
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io/fs"
	"log"
	"os"
	"runtime/debug"
	"strconv"
	"sync"
	"time"

	"github.com/remeh/sizedwaitgroup"
)

var (
	DBWriteLock sync.Mutex
	DBLock      sync.RWMutex
	LID_TOP     uint32
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

	for i := range Clusters {
		WriteCluster(i)
	}
	endTime := time.Now()
	cwlog.DoLog("DB Write Complete, took: " + endTime.Sub(startTime).String())

}

// Size of all fields, sans version header
const RecordSize = 38

func WriteCluster(i int) {
	//startTime := time.Now()

	var buf [(RecordSize * (cons.ClusterSize)) + 2]byte
	var b int64
	binary.LittleEndian.PutUint16(buf[b:], 1) //version number
	b += 2

	for gi, g := range Clusters[i].Guilds {

		if g == nil {
			break
		}
		Clusters[i].Guilds[gi].Lock.RLock()
		binary.LittleEndian.PutUint32(buf[b:], g.LID)
		b += 4
		binary.LittleEndian.PutUint64(buf[b:], g.Customer)
		b += 8
		binary.LittleEndian.PutUint64(buf[b:], g.Guild)
		b += 8
		binary.LittleEndian.PutUint64(buf[b:], g.Added)
		b += 8
		binary.LittleEndian.PutUint64(buf[b:], g.Modified)
		b += 8
		binary.LittleEndian.PutUint16(buf[b:], g.Donator)
		b += 2

		Clusters[i].Guilds[gi].Lock.RUnlock()
	}
	name := fmt.Sprintf("data/db/cluster-%v.dat", i+1)
	err := os.WriteFile(name+".tmp", buf[0:b], 0644)
	if err != nil && err != fs.ErrNotExist {
		cwlog.DoLog(err.Error())
		return
	}
	err = os.Rename(name+".tmp", name)

	if err != nil {
		cwlog.DoLog("WriteCluster: Couldn't rename file: " + name)
		return
	}

	//endTime := time.Now()
	//cwlog.DoLog("Cluster-" + strconv.FormatInt(int64(i+1), 10) + " write, took: " + endTime.Sub(startTime).String() + ", Wrote: " + strconv.FormatInt(b, 10) + "b")
}

func ReadAllClusters() {

	wg := sizedwaitgroup.New(ThreadCount)

	startTime := time.Now()
	for x := 0; x < cons.TSize/cons.ClusterSize && x < cons.NumClusters; x++ {
		wg.Add()
		go func(x int) {
			ReadCluster(int64(x))
			wg.Done()
		}(x)
	}
	wg.Wait()
	endTime := time.Now()

	cwlog.DoLog("Read all clusters, took: " + endTime.Sub(startTime).String())
}

func ReadCluster(i int64) {
	startTime := time.Now()

	name := fmt.Sprintf("data/db/cluster-%v.dat", i+1)
	data, err := os.ReadFile(name)
	if err != nil {
		cwlog.DoLog(err.Error())
		return
	}

	dataLen := int64(len(data))
	var b int64
	var gi int64

	version := binary.LittleEndian.Uint16(data[b:])
	b += 2
	if version == 1 {
		for b < dataLen {
			if (dataLen-b)-RecordSize < 0 {
				cwlog.DoLog("Invalid cluster data, stopping.")
				break
			}
			var g *GuildData = Clusters[i].Guilds[gi]
			if g == nil {
				g = &GuildData{}
			}

			g.LID = binary.LittleEndian.Uint32(data[b:])
			b += 4
			g.Customer = binary.LittleEndian.Uint64(data[b:])
			b += 8
			g.Guild = binary.LittleEndian.Uint64(data[b:])
			b += 8
			g.Added = binary.LittleEndian.Uint64(data[b:])
			b += 8
			g.Modified = binary.LittleEndian.Uint64(data[b:])
			b += 8
			g.Donator = binary.LittleEndian.Uint16(data[b:])
			b += 2

			Clusters[i].Guilds[gi] = g
			gi++
		}
	} else {
		cwlog.DoLog("Invalid cluster version.")
		return
	}

	endTime := time.Now()
	cwlog.DoLog("Cluster-" + strconv.FormatInt(int64(i+1), 10) + " read, took: " + endTime.Sub(startTime).String() + ", Read: " + strconv.FormatInt(b, 10) + "b")
}

func UpdateGuildLookup() {
	startTime := time.Now()
	cwlog.DoLog("Updating guild lookup map.")

	count := 0
	for ci := 0; ci < cons.NumClusters; ci++ {
		for gi := 0; gi < cons.ClusterSize; gi++ {
			if Clusters[ci].Guilds[gi] == nil {
				break
			}

			gid := Clusters[ci].Guilds[gi].Guild
			if GuildLookup[gid] == nil {
				GuildLookupLock.Lock()
				GuildLookup[gid] = Clusters[ci].Guilds[gi]
				count++
				GuildLookupLock.Unlock()
			} else {
				cwlog.DoLog("*** GUILD-ID ALREADY OCCUPIED ***")
			}
		}
	}

	debug.FreeOSMemory()
	endTime := time.Now()

	buf := fmt.Sprintf("guilds: %v", count)
	cwlog.DoLog(buf)
	cwlog.DoLog("Guild lookup map update, took: " + endTime.Sub(startTime).String())
}

func GuildLookupRead(i uint64) *GuildData {
	GuildLookupLock.RLock()
	g := GuildLookup[i]
	GuildLookupLock.RUnlock()
	return g
}
