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
const RecordSize = 32

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
		binary.LittleEndian.PutUint32(buf[b:], g.Added)
		b += 4
		binary.LittleEndian.PutUint32(buf[b:], g.Modified)
		b += 4
		binary.LittleEndian.PutUint16(buf[b:], g.Donator)
		b += 2
		buf[b] = cons.RecordEnd
		b += 1

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
			g.Added = binary.LittleEndian.Uint32(data[b:])
			b += 4
			g.Modified = binary.LittleEndian.Uint32(data[b:])
			b += 4
			g.Donator = binary.LittleEndian.Uint16(data[b:])
			b += 2
			end := data[b]
			b += 1

			if end == cons.RecordEnd {
				Clusters[i].Guilds[gi] = g
				gi++
			} else {
				buf := fmt.Sprintf("ReadCluster: %v: %v: INVALID RECORD!", name, gi)
				cwlog.DoLog(buf)
				return
			}
		}
	} else {
		cwlog.DoLog("Invalid cluster version.")
		return
	}

	endTime := time.Now()
	cwlog.DoLog("Cluster-" + strconv.FormatInt(int64(i+1), 10) + " read, took: " + endTime.Sub(startTime).String() + ", Read: " + strconv.FormatInt(b, 10) + "b")
}

func AppendCluster(guild *GuildData, cid uint32, gid uint32) {

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

func AddGuild(guildid uint64) {

	LID_TOP++
	cid := LID_TOP % cons.NumClusters
	gid := (LID_TOP % cons.ClusterSize) / cons.NumClusters

	tNow := NowToCompact()
	Clusters[cid].Guilds[gid] = &GuildData{LID: LID_TOP, Guild: guildid, Added: uint32(tNow), Modified: uint32(tNow)}
	UpdateGuildLookup()
}

func UpdateGuild(guild *GuildData) {
	if guild == nil {
		return
	}
	lid := guild.LID
	cid := lid % cons.NumClusters
	gid := (lid % cons.ClusterSize) / cons.NumClusters

	guildData := Clusters[cid].Guilds[gid]
	AppendCluster(guildData, cid, gid)
}

// Give current time in compact format
func NowToCompact() uint32 {
	tNow := time.Now().UTC().Unix()
	return uint32(tNow - cons.RoleKeeperEpoch)
}

// Compact format to unix time
func CompactToUnix(input uint32) uint64 {
	return uint64(input) + cons.RoleKeeperEpoch
}

func UnixToCompact(input uint64) uint32 {
	return uint32(input - cons.RoleKeeperEpoch)
}
