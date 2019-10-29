// +build linux

package synproxy

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Synproxy struct {
	// Synproxy stats filename (proc filesystem)
	statFile string
}

func (k *Synproxy) Description() string {
	return fmt.Sprintf("Get synproxy statistics from %s", k.statFile)
}

func (k *Synproxy) SampleConfig() string {
	return ""
}

func (k *Synproxy) Gather(acc telegraf.Accumulator) error {
	data, err := k.getSynproxyStat()
	if err != nil {
		return err
	}

	acc.AddCounter("synproxy", data, map[string]string{})
	return nil
}

func (k *Synproxy) getSynproxyStat() (map[string]interface{}, error) {
	var hname []string
	fields := make(map[string]interface{})

	// Open synproxy file in proc filesystem
	file, err := os.Open(k.statFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read result
	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		line := scanner.Text()
		// Parse fields separated by whitespace
		dataFields := strings.Fields(line)
		for _, val := range dataFields {
			hname = append(hname, val)
			fields[val] = uint32(0)
		}
	}
	if len(fields) == 0 {
		return nil, fmt.Errorf("invalid data")
	}
	for scanner.Scan() {
		line := scanner.Text()
		// Parse fields separated by whitespace
		dataFields := strings.Fields(line)
		// If number of data fields do not match number of header fields
		if len(dataFields) != len(fields) {
			return nil, fmt.Errorf("invalid number of columns in data, expected %d found %d", len(fields),
				len(dataFields))
		}
		for i, val := range dataFields {
			// Convert from hexstring to int32
			x, err := strconv.ParseUint(val, 16, 32)
			// If field is not a valid hexstring
			if err != nil {
				return nil, fmt.Errorf("invalid value '%s' found", val)
			}
			fields[hname[i]] = fields[hname[i]].(uint32) + uint32(x)
		}
	}
	return fields, nil
}

func GetHostProc() string {
	procPath := "/proc"
	if os.Getenv("HOST_PROC") != "" {
		procPath = os.Getenv("HOST_PROC")
	}
	return procPath
}

func init() {
	inputs.Add("synproxy", func() telegraf.Input {
		return &Synproxy{
			statFile: path.Join(GetHostProc(), "/net/stat/synproxy"),
		}
	})
}
