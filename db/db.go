package db

import (
	"RoleKeeper/cons"
	"RoleKeeper/cwlog"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/remeh/sizedwaitgroup"
)

func WriteLIDTop() {
	buf := fmt.Sprintf("LIDTop: %v", LID_TOP)
	err := os.WriteFile("data/db/"+cons.LIDTopFile+".tmp", []byte(buf), 0644)

	if err != nil && err != fs.ErrNotExist {
		cwlog.DoLog(err.Error())
		return
	}
	err = os.Rename("data/db/"+cons.LIDTopFile+".tmp", "data/db/"+cons.LIDTopFile)

	if err != nil {
		cwlog.DoLog("WriteLIDTop: Couldn't rename file: " + cons.LIDTopFile)
		return
	}
}

func ReadLIDTop() int64 {
	data, err := os.ReadFile("data/db/" + cons.LIDTopFile)
	if err != nil {
		cwlog.DoLog("Unable to read LIDTop file, exiting")
		os.Exit(1)
		return -1
	}
	splitData := strings.Split(string(data), ":")

	if len(splitData) > 1 {
		valString := splitData[1]
		cleanVal := strings.TrimSuffix(valString, "\n")
		cleanVal = strings.TrimPrefix(cleanVal, " ")
		intVal, err := strconv.ParseUint(cleanVal, 10, 32)

		if err != nil {
			cwlog.DoLog("Unable to parse LIDTop value, exiting")
			os.Exit(1)
			return -1
		}

		return int64(intVal)
	}
	return -1
}

func WriteAllCluster() {

	startTime := time.Now()

	for i := 0; i < cons.NumClusters; i++ {
		WriteCluster(i)
	}
	endTime := time.Now()
	cwlog.DoLog("DB Write Complete, took: " + endTime.Sub(startTime).String())

}

const VERSION_size = 2

const LID_size = 4
const SNOWFLAKE_size = 8
const ADDED_size = 4
const DONOR_size = 1
const NUMROLE_size = 2

const MAXROLES = 0xFFFF
const MAXROLE_size = SNOWFLAKE_size * MAXROLES

const staticSize = LID_size +
	SNOWFLAKE_size +
	ADDED_size +
	DONOR_size +
	NUMROLE_size

const clusterSize = (VERSION_size + ((staticSize + MAXROLE_size) * (cons.MaxGuilds / cons.NumClusters)))

func WriteCluster(i int) {

	var buf [clusterSize]byte

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

		numRoles := uint16(len(g.Roles))
		binary.LittleEndian.PutUint16(buf[b:], numRoles)
		b += 2

		for c, role := range g.Roles {
			if c < MAXROLES {
				binary.LittleEndian.PutUint64(buf[b:], role.ID)
				b += 8
			}
		}

		Database[x].Lock.RUnlock()
	}
	/* Don't write empty files */
	if b > VERSION_size {
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

	lid := ReadLIDTop()
	if lid != int64(LID_TOP) {
		cwlog.DoLog(fmt.Sprintf("ERROR: LIDTOPFile: %v, LIDTOP: %v\nSTOPPING for 5 minutes!\n", lid, LID_TOP))
		time.Sleep(time.Second * 300)
		os.Exit(1)
		return
	}
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

	if (dataLen) < VERSION_size+staticSize {
		fmt.Println("cluster size too small:", dataLen, " bytes of ", VERSION_size+staticSize)
		return
	}

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

			numRoles := binary.LittleEndian.Uint16(data[b:])
			b += 2

			if numRoles == 0 {
				if g.LID >= cons.MaxGuilds {
					cwlog.DoLog("LID larger than maxguild.")
					break
				}
				LID_TOP++
				Database[LID_TOP] = g
				/* Found a role instead of record end */
			} else {
				roleData := []RoleData{}
				for i := uint16(1); i <= numRoles; i++ {
					roleID := binary.LittleEndian.Uint64(data[b:])
					b += 8

					roleData = append(roleData, RoleData{ID: roleID})
				}
				g.Roles = roleData
				LID_TOP++
				Database[LID_TOP] = g
			}
		}
	} else {
		cwlog.DoLog("Invalid cluster version.")
		return
	}

	endTime := time.Now()
	if b > 2 {
		cwlog.DoLog("Cluster-" + strconv.FormatInt(int64(i+1), 10) + " read, took: " + endTime.Sub(startTime).String() + ", Read: " + strconv.FormatInt(b, 10) + "b")
	}
}

func AppendCluster(guild *GuildData, cid uint32, gid uint32) {

}

func UpdateGuildLookup() {
	startTime := time.Now()

	GuildLookupLock.Lock()
	var x uint32
	for x = 1; x <= LID_TOP; x++ {
		if Database[x] != nil {
			GuildLookup[Database[x].Guild] = Database[x]
		}
	}
	GuildLookupLock.Unlock()

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
	val, err := SnowflakeToInt(i)
	if err == nil {
		g := GuildLookup[val]
		GuildLookupLock.RUnlock()
		return g
	}
	GuildLookupLock.RUnlock()
	return nil
}

// Does not need a lock, this is pre-allocated.
func AddGuild(guildid uint64) {
	cwlog.DoLog(fmt.Sprintf("AddGuild: %v", guildid))

	LID_TOP++
	tNow := NowToCompact()
	newGuild := GuildData{LID: LID_TOP, Guild: guildid, Added: uint32(tNow), Modified: uint32(tNow), Donator: 0}
	Database[LID_TOP] = &newGuild

	WriteLIDTop()
	UpdateGuildLookup()
}

func DumpGuilds() {

	fo, err := os.Create("data/" + cons.DumpName)
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
	enc.SetIndent("", "\t")
	if err := enc.Encode(GuildLookup); err != nil {
		cwlog.DoLog("DumpGuilds: enc.Encode failure")
		GuildLookupLock.RUnlock()
		return
	}
	GuildLookupLock.RUnlock()

	nfilename := "data/" + cons.DumpName + ".tmp"
	//compBuf := compressZip(outbuf.Bytes())
	err = os.WriteFile(nfilename, outbuf.Bytes(), 0644)

	if err != nil {
		cwlog.DoLog("DumpGuilds: Couldn't write db temp file.")
		return
	}

	oldName := nfilename
	newName := "data/" + cons.DumpName
	err = os.Rename(oldName, newName)

	if err != nil {
		cwlog.DoLog("DumpGuilds: Couldn't rename db temp file.")
		return
	}

	cwlog.DoLog("DumpGuilds: Complete!")
}
