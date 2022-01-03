//go:build linux
// +build linux

package synproxy

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
)

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
