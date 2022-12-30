package db

import (
	"RoleKeeper/cons"
	"RoleKeeper/cwlog"
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/remeh/sizedwaitgroup"
)

type RoleData struct {
	Name string
	ID   uint64
}

type GuildData struct {
	//Name type bytes
	LID      uint32 `json:"l,omitempty"` //4
	Customer uint64 `json:"c,omitempty"` //8
	Guild    uint64 `json:"-"`           //8 --Already in JSON as KEY
	Added    uint32 `json:"a,omitempty"` //4
	Modified uint32 `json:"m,omitempty"` //4

	Donator uint8 `json:"d,omitempty"` //1

	/* Not on disk */
	Roles []RoleData   `json:"-"`
	Lock  sync.RWMutex `json:"-"`
}

var (
	LID_TOP         uint32 = 0
	GuildLookup     map[uint64]*GuildData
	GuildLookupLock sync.RWMutex
	ThreadCount     int

	Database [cons.MaxGuilds]*GuildData
)

func IntToID(id uint64) string {
	strId := fmt.Sprintf("%v", id)
	return strId
}

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

func WriteLIDTop() {
	buf := fmt.Sprintf("LIDTop: %v", LID_TOP)
	err := os.WriteFile("data/"+cons.LIDTopFile+".tmp", []byte(buf), 0644)

	if err != nil && err != fs.ErrNotExist {
		cwlog.DoLog(err.Error())
		return
	}
	err = os.Rename("data/"+cons.LIDTopFile+".tmp", "data/"+cons.LIDTopFile)

	if err != nil {
		cwlog.DoLog("WriteLIDTop: Couldn't rename file: " + cons.LIDTopFile)
		return
	}
}

func WriteAllCluster() {

	startTime := time.Now()

	for i := 0; i < cons.NumClusters; i++ {
		WriteCluster(i)
	}
	endTime := time.Now()
	cwlog.DoLog("DB Write Complete, took: " + endTime.Sub(startTime).String())

}

// Size of all fields, sans version header
const v1RecordSize = 19

func WriteCluster(i int) {

	var buf [(v1RecordSize * (cons.MaxGuilds / cons.NumClusters)) + 2]byte
	var b int64
	binary.LittleEndian.PutUint16(buf[b:], 1) //version number
	b += 2

	start := i*(cons.MaxGuilds/cons.NumClusters) + 1
	end := start + (cons.MaxGuilds / cons.NumClusters)
	for x := start; x < end; x++ {

		g := Database[x]
		if g == nil {
			break
		}

		Database[x].Lock.RLock()
		binary.LittleEndian.PutUint32(buf[b:], g.LID)
		b += 4
		binary.LittleEndian.PutUint64(buf[b:], g.Guild)
		b += 8
		binary.LittleEndian.PutUint32(buf[b:], g.Added)
		b += 4
		buf[b] = g.Donator
		b += 1
		buf[b] = cons.RecordEnd
		b += 1

		Database[x].Lock.RUnlock()
	}
	/* Don't write empty files */
	if b > 2 {
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
	}
}

/* TODO: Read whole folder for cluster files, warn if LID_TOP does not match */
func ReadAllClusters() {

	wg := sizedwaitgroup.New(ThreadCount)

	startTime := time.Now()
	for x := 0; x < cons.NumClusters; x++ {
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
		return
	}

	dataLen := int64(len(data))
	var b int64

	version := binary.LittleEndian.Uint16(data[b:])
	b += 2
	if version == 1 {
		for b < dataLen {
			g := new(GuildData)

			g.LID = binary.LittleEndian.Uint32(data[b:])
			b += 4
			g.Guild = binary.LittleEndian.Uint64(data[b:])
			b += 8
			g.Added = binary.LittleEndian.Uint32(data[b:])
			b += 4
			g.Donator = data[b]
			b += 1
			end := data[b]
			b += 1

			if end == cons.RecordEnd {
				if g.LID >= cons.MaxGuilds {
					cwlog.DoLog("LID larger than maxguild.")
					continue
				}
				if g.LID > LID_TOP {
					LID_TOP = g.LID
				}
				Database[g.LID] = g
			} else {
				buf := fmt.Sprintf("ReadCluster: %v: %v: INVALID RECORD!", name, g.LID)
				cwlog.DoLog(buf)
				return
			}
		}
	} else {
		cwlog.DoLog("Invalid cluster version.")
		return
	}

	endTime := time.Now()
	if b > 2 && 1 == 2 {
		cwlog.DoLog("Cluster-" + strconv.FormatInt(int64(i+1), 10) + " read, took: " + endTime.Sub(startTime).String() + ", Read: " + strconv.FormatInt(b, 10) + "b")
	}
}

func AppendCluster(guild *GuildData, cid uint32, gid uint32) {

}

func UpdateGuildLookup() {
	startTime := time.Now()

	var x uint32
	for x = 1; x <= LID_TOP; x++ {
		if Database[x] != nil {
			GuildLookup[Database[x].Guild] = Database[x]
		}
	}
	endTime := time.Now()
	cwlog.DoLog("Guild lookup map update, took: " + endTime.Sub(startTime).String())

	WriteAllCluster()
	DumpGuilds()
}

func GuildLookupRead(i uint64) *GuildData {
	GuildLookupLock.RLock()
	g := GuildLookup[i]
	GuildLookupLock.RUnlock()
	return g
}

func GuildLookupReadString(i string) *GuildData {
	GuildLookupLock.RLock()
	val, err := GuildStrToInt(i)
	if err == nil {
		g := GuildLookup[val]
		return g
	}
	GuildLookupLock.RUnlock()
	return nil
}

func GuildStrToInt(i string) (uint64, error) {
	return strconv.ParseUint(i, 10, 64)
}

func AddGuild(guildid uint64) {
	cwlog.DoLog(fmt.Sprintf("AddGuild: %v", guildid))

	LID_TOP++
	tNow := NowToCompact()
	newGuild := GuildData{LID: LID_TOP, Guild: guildid, Added: uint32(tNow), Modified: uint32(tNow), Donator: 0}
	Database[LID_TOP] = &newGuild

	WriteLIDTop()
	UpdateGuildLookup()
}

// Current time in compact format
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

func DumpGuilds() {

	fo, err := os.Create(cons.DumpName)
	if err != nil {
		cwlog.DoLog("Couldn't open db file, skipping...")
		return
	}
	/*  close fo on exit and check for its returned error */
	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()

	GuildLookupLock.RLock()

	cwlog.DoLog("DumpGuilds: Writing guilds...")

	outbuf := new(bytes.Buffer)
	enc := json.NewEncoder(outbuf)
	if err := enc.Encode(GuildLookup); err != nil {
		cwlog.DoLog("DumpGuilds: enc.Encode failure")
		return
	}
	GuildLookupLock.RUnlock()

	nfilename := cons.DumpName + ".tmp"
	//compBuf := compressZip(outbuf.Bytes())
	err = os.WriteFile(nfilename, outbuf.Bytes(), 0644)

	if err != nil {
		cwlog.DoLog("DumpGuilds: Couldn't write db temp file.")
		return
	}

	oldName := nfilename
	newName := cons.DumpName
	err = os.Rename(oldName, newName)

	if err != nil {
		cwlog.DoLog("DumpGuilds: Couldn't rename db temp file.")
		return
	}

	cwlog.DoLog("DumpGuilds: Complete!")
}
