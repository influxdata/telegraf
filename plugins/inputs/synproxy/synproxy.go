//go:build linux

//go:generate ../../../tools/readme_config_includer/generator
package synproxy

import (
	"bufio"
	_ "embed"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// DO NOT REMOVE THE NEXT TWO LINES! This is required to embed the sampleConfig data.
//
//go:embed sample.conf
var sampleConfig string

type Synproxy struct {
	Log telegraf.Logger `toml:"-"`

	// Synproxy stats filename (proc filesystem)
	statFile string
}

func (*Synproxy) SampleConfig() string {
	return sampleConfig
}

func (k *Synproxy) Gather(acc telegraf.Accumulator) error {
	data, err := k.getSynproxyStat()
	if err != nil {
		return err
	}

	acc.AddCounter("synproxy", data, map[string]string{})
	return nil
}

func inSlice(haystack []string, needle string) bool {
	for _, val := range haystack {
		if needle == val {
			return true
		}
	}
	return false
}

func (k *Synproxy) getSynproxyStat() (map[string]interface{}, error) {
	var hname []string
	counters := []string{"entries", "syn_received", "cookie_invalid", "cookie_valid", "cookie_retrans", "conn_reopened"}
	fields := make(map[string]interface{})

	// Open synproxy file in proc filesystem
	file, err := os.Open(k.statFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Initialise expected fields
	for _, val := range counters {
		fields[val] = uint32(0)
	}

	scanner := bufio.NewScanner(file)
	// Read header row
	if scanner.Scan() {
		line := scanner.Text()
		// Parse fields separated by whitespace
		dataFields := strings.Fields(line)
		for _, val := range dataFields {
			if !inSlice(counters, val) {
				val = ""
			}
			hname = append(hname, val)
		}
	}
	if len(hname) == 0 {
		return nil, fmt.Errorf("invalid data")
	}
	// Read data rows
	for scanner.Scan() {
		line := scanner.Text()
		// Parse fields separated by whitespace
		dataFields := strings.Fields(line)
		// If number of data fields do not match number of header fields
		if len(dataFields) != len(hname) {
			return nil, fmt.Errorf("invalid number of columns in data, expected %d found %d", len(hname),
				len(dataFields))
		}
		for i, val := range dataFields {
			// Convert from hexstring to int32
			x, err := strconv.ParseUint(val, 16, 32)
			// If field is not a valid hexstring
			if err != nil {
				return nil, fmt.Errorf("invalid value '%s' found", val)
			}
			if hname[i] != "" {
				fields[hname[i]] = fields[hname[i]].(uint32) + uint32(x)
			}
		}
	}
	return fields, nil
}

func getHostProc() string {
	procPath := "/proc"
	if os.Getenv("HOST_PROC") != "" {
		procPath = os.Getenv("HOST_PROC")
	}
	return procPath
}

func init() {
	inputs.Add("synproxy", func() telegraf.Input {
		return &Synproxy{
			statFile: path.Join(getHostProc(), "/net/stat/synproxy"),
		}
	})
}
