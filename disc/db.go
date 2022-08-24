package disc

import (
	"RoleKeeper/cons"
	"RoleKeeper/rclog"
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/remeh/sizedwaitgroup"
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

	wg := sizedwaitgroup.New(ThreadCount)

	for i, c := range Clusters {
		if c == nil {
			break
		}
		wg.Add()
		go func(i int) {
			WriteCluster(i)
			wg.Done()
		}(i)
	}
	wg.Wait()
	endTime := time.Now()
	rclog.DoLog("DB Write Complete, took: " + endTime.Sub(startTime).String())

}

func WriteCluster(i int) {
	startTime := time.Now()

	var buf [(RecordSize * (cons.ClusterSize)) + 2]byte
	var b int64
	binary.BigEndian.PutUint16(buf[b:], 1) //version number
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
		binary.LittleEndian.PutUint16(buf[b:], uint16(g.Donator))
		b += 2
		binary.LittleEndian.PutUint16(buf[b:], uint16(g.Premium))
		b += 2
		Clusters[i].Guilds[gi].Lock.RUnlock()
	}
	name := fmt.Sprintf("db/cluster-%v.dat", i+1)
	err := os.WriteFile(name, buf[0:b], 0644)
	if err != nil && err != fs.ErrNotExist {
		rclog.DoLog(err.Error())
		return
	}

	endTime := time.Now()
	rclog.DoLog("Cluster-" + strconv.FormatInt(int64(i+1), 10) + " write, took: " + endTime.Sub(startTime).String() + ", Wrote: " + strconv.FormatInt(b, 10) + "b")
}

func ReadAllClusters() {

	wg := sizedwaitgroup.New(ThreadCount)

	startTime := time.Now()
	for x := 0; x < cons.TSize/cons.ClusterSize && x < cons.MaxClusters; x++ {
		wg.Add()
		go func(x int) {
			ReadCluster(int64(x))
			wg.Done()
		}(x)
	}
	wg.Wait()
	endTime := time.Now()

	rclog.DoLog("Read all clusters, took: " + endTime.Sub(startTime).String())
}

func ReadCluster(i int64) {
	startTime := time.Now()

	name := fmt.Sprintf("db/cluster-%v.dat", i+1)
	data, err := os.ReadFile(name)
	if err != nil {
		rclog.DoLog(err.Error())
		return
	}

	dataLen := int64(len(data))
	var b int64
	var gi int64

	if Clusters[i] == nil {
		c := ClusterData{}
		Clusters[i] = &c
		rclog.DoLog("New cluster: " + strconv.FormatInt(i+1, 10))
	}

	version := binary.BigEndian.Uint16(data[b:])
	b += 2
	if version == 1 {
		for b < dataLen {
			if (dataLen-b)-RecordSize < 0 {
				rclog.DoLog("Invalid cluster data, stopping.")
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
			g.Donator = uint8(binary.LittleEndian.Uint16(data[b:]))
			b += 2
			g.Premium = uint8(binary.LittleEndian.Uint16(data[b:]))
			b += 2

			Clusters[i].Guilds[gi] = g
			gi++
		}
	} else {
		rclog.DoLog("Invalid cluster version.")
		return
	}

	endTime := time.Now()
	rclog.DoLog("Cluster-" + strconv.FormatInt(int64(i+1), 10) + " read, took: " + endTime.Sub(startTime).String() + ", Read: " + strconv.FormatInt(b, 10) + "b")
}

func UpdateGuildLookup() {
	startTime := time.Now()
	rclog.DoLog("Updating guild lookup map.")

	for ci, c := range Clusters {
		if c == nil {
			break
		}
		for gi, g := range c.Guilds {
			if g == nil {
				break
			}
			if GuildLookup[g.Guild] == nil {
				GuildLookupLock.Lock()
				GuildLookup[g.Guild] = Clusters[ci].Guilds[gi]
				GuildLookupLock.Unlock()
			}
		}
	}

	endTime := time.Now()
	rclog.DoLog("Guild lookup map update, took: " + endTime.Sub(startTime).String())
}

func GuildLookupRead(i uint64) *GuildData {
	GuildLookupLock.RLock()
	g := GuildLookup[i]
	GuildLookupLock.RUnlock()
	return g
}
