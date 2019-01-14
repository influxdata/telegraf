package linux_fscache

import (
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type FSCache struct {
	statFile string
}

func (f *FSCache) Description() string {
	return "Get FS-Cache statistics from /proc/fs/fscache/stats"
}

func (f *FSCache) SampleConfig() string { return "" }

func (f *FSCache) Gather(acc telegraf.Accumulator) error {

	s, err := os.Stat(f.statFile)
	if err != nil {
		return err
	} else if s.Size() == 0 {
		return nil
	}

	data, err := ioutil.ReadFile(f.statFile)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) <= 1 {
		return nil
	}

	fields := make(map[string]interface{})
	for _, line := range lines[1:] {
		parts := strings.Split(line, ":")
		if len(parts) >= 2 {
			prefix := strings.TrimSpace(parts[0])
			subparts := strings.Fields(parts[1])
			for _, subpart := range subparts {
				values := strings.Split(strings.TrimSpace(subpart), "=")
				if len(values) == 2 {
					v, err := strconv.ParseInt(values[1], 10, 64)
					if err != nil {
						return err
					}
					sn := internal.SnakeCase(prefix+"_"+values[0])
					fields[sn] = int64(v)
				}
			}
		}
	}

	acc.AddCounter("linux_fscache", fields, nil)

	return nil
}

func init() {
	inputs.Add("linux_fscache", func() telegraf.Input {
		return &FSCache{
			statFile: "/proc/fs/fscache/stats",
		}
	})
}
