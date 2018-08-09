package meminfo

import (
	"bufio"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"os"
	"strconv"
	"strings"
)

type MemStats struct {
	Fields map[string]interface{}
}

func (s *MemStats) Description() string {
	return "Memory Information from /proc/meminfo"
}

func (s *MemStats) SampleConfig() string {
	return `
  [inputs.meminfo]
	This plugin requires no settings
`
}

func (s *MemStats) Gather(acc telegraf.Accumulator) error {
	data, err := os.Open("/proc/meminfo")
	if err != nil {
		panic(err)
	}
	defer data.Close()
	fileScanner := bufio.NewScanner(data)
	for fileScanner.Scan() {
		stat := strings.Fields(fileScanner.Text())
		name := stat[0][:len(stat[0])-1]
		value, err := strconv.ParseInt(stat[1], 10, 64)
		if err != nil {
			panic(err)
		}
		if len(stat) > 2 {
			// If we have "kB" in the line, we should make it bytes
			value = value * 1024
		}
		s.Fields[name] = value
	}
	acc.AddGauge("meminfo", s.Fields, nil)

	return nil
}

func init() {
	inputs.Add("meminfo", func() telegraf.Input {
		return &MemStats{Fields: make(map[string]interface{})}
	})
}
