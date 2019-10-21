// +build linux

package synproxy

import (
	"bufio"
	"fmt"
	"os"
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
		return nil, fmt.Errorf("synproxy: %s does not exist!", k.statFile)
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
		return nil, fmt.Errorf("synproxy: Invalid data!")
	}
	for scanner.Scan() {
		line := scanner.Text()
		// Parse fields separated by whitespace
		dataFields := strings.Fields(line)
		for i, val := range dataFields {
			// Convert from hexstring to int32
			x, err := strconv.ParseUint(val, 16, 32)
			// If field is not a valid hexstring
			if err != nil {
				return nil, fmt.Errorf("synproxy: Invalid value '%s' found!", val)
			// If index is out of boundary
			} else if i >= len(fields) {
				return nil, fmt.Errorf("synproxy: Value '%s' out of column boundary!", val)
			// If field is a valid hexstring and index not out of boundary
			} else {
				fields[hname[i]] = fields[hname[i]].(uint32) + uint32(x)
			}
		}
	}
	return fields, nil
}

func init() {
	inputs.Add("synproxy", func() telegraf.Input {
		return &Synproxy{
			statFile: "/proc/net/stat/synproxy",
		}
	})
}
