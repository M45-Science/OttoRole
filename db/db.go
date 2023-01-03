package db

import (
	"RoleKeeper/cons"
	"RoleKeeper/cwlog"
	"RoleKeeper/disc"
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/remeh/sizedwaitgroup"
	"github.com/sasha-s/go-deadlock"
)

type RoleData struct {
	Name string `json:"n,omitempty"`
	ID   uint64 `json:"i"`
}

type GuildData struct {
	//Name type bytes
	LID      uint32     `json:"l,omitempty"`
	Customer uint64     `json:"c,omitempty"`
	Guild    uint64     `json:"-"` //Already in JSON as KEY
	Added    uint32     `json:"a,omitempty"`
	Modified uint32     `json:"m,omitempty"`
	Donator  uint8      `json:"d,omitempty"`
	Roles    []RoleData `json:"r,omitempty"`

	/* Not on disk */
	Lock deadlock.RWMutex `json:"-"`
}

var (
	LID_TOP         uint32 = 0
	GuildLookup     map[uint64]*GuildData
	GuildLookupLock deadlock.RWMutex
	ThreadCount     int

	Database [cons.MaxGuilds]*GuildData
)

func GuildRoleCreate(s *discordgo.Session, role *discordgo.GuildRoleCreate) {

	cwlog.DoLog("Role created.")
}
func GuildRoleUpdate(s *discordgo.Session, role *discordgo.GuildRoleUpdate) {

	cwlog.DoLog("Role modified.")
}
func GuildRoleDelete(s *discordgo.Session, role *discordgo.GuildRoleDelete) {

	cwlog.DoLog("Role deleted.")
}

func IntToSnowflake(id uint64) string {
	strId := fmt.Sprintf("%v", id)
	return strId
}

func SnowflakeToInt(i string) (uint64, error) {
	return strconv.ParseUint(i, 10, 64)
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

func LookupRoleNames(s *discordgo.Session, guildData *GuildData) {
	startTime := time.Now()
	count := 0

	//Process all guilds
	if guildData == nil {
		GuildLookupLock.Lock()
		for gpos, guild := range GuildLookup {
			time.Sleep(time.Millisecond) //Let other processes run
			guild.Lock.Lock()

			for rpos, role := range guild.Roles {
				if role.Name == "" {
					roleList := disc.GetGuildRoles(s, IntToSnowflake(guild.Guild))
					for _, discRole := range roleList {
						discRoleID, err := SnowflakeToInt(discRole.ID)
						if err == nil {
							if role.ID == discRoleID {
								if discRole.Name != "" {

									GuildLookup[gpos].Roles[rpos].Name = discRole.Name
									count++
								}
							}
						}
					}
				}
			}

			guild.Lock.Unlock()
		}
		GuildLookupLock.Unlock()
		buf := fmt.Sprintf("Added %v role names in %v.", count, time.Since(startTime).String())
		cwlog.DoLog(buf)
	} else { //Process a specific guild

		guildData.Lock.Lock()
		for rpos, role := range guildData.Roles {
			discGuild, err := s.Guild(IntToSnowflake(guildData.Guild))
			if err == nil {
				for _, discRole := range discGuild.Roles {
					if IntToSnowflake(role.ID) == discRole.ID {
						guildData.Roles[rpos].Name = discRole.Name
					}
				}
			}
		}
		guildData.Lock.Unlock()
	}
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
					continue
				}
				LID_TOP++
				Database[LID_TOP] = g
				break
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
				break
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
		GuildLookupLock.RUnlock()
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
