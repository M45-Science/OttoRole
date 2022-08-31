package cfg

import (
	"RoleKeeper/cons"
	"RoleKeeper/cwlog"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
)

var Config serverConfig

type serverConfig struct {
	Token  string
	App    string
	Domain string
}

func WriteCfg() bool {
	tempPath := cons.ConfigFile + ".tmp"
	finalPath := cons.ConfigFile

	outbuf := new(bytes.Buffer)
	enc := json.NewEncoder(outbuf)
	enc.SetIndent("", "\t")

	if err := enc.Encode(Config); err != nil {
		cwlog.DoLog("WriteCfg: enc.Encode failure")
		return false
	}

	_, err := os.Create(tempPath)

	if err != nil {
		cwlog.DoLog("WriteCfg: os.Create failure")
		return false
	}

	err = ioutil.WriteFile(tempPath, outbuf.Bytes(), 0644)

	if err != nil {
		cwlog.DoLog("WriteCfg: WriteFile failure")
	}

	err = os.Rename(tempPath, finalPath)

	if err != nil {
		cwlog.DoLog("WriteCfg: Couldn't rename cfg file.")
		return false
	}

	return true
}

func ReadCfg() bool {

	_, err := os.Stat(cons.ConfigFile)
	notfound := os.IsNotExist(err)

	if notfound {
		cwlog.DoLog("ReadCfg: os.Stat failed, empty config generated.")
		return true
	} else { /* Otherwise just read in the config */
		file, err := ioutil.ReadFile(cons.ConfigFile)

		if file != nil && err == nil {
			newcfg := serverConfig{}

			err := json.Unmarshal([]byte(file), &newcfg)
			if err != nil {
				cwlog.DoLog("ReadCfg: Unmarshal failure")
				cwlog.DoLog(err.Error())
				return false
			}

			Config = newcfg
			return true
		} else {
			cwlog.DoLog("ReadCfg: ReadFile failure")
			return false
		}
	}
}
