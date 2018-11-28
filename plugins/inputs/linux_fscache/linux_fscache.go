package linux_fscache

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
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
		return fmt.Errorf("fscache: %s zero length!", f.statFile)
	}

	data, err := ioutil.ReadFile(f.statFile)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	fields := make(map[string]interface{})
	for i := 1; i < len(lines); i++ {
		parts := strings.Split(lines[i], ":")
		prefix := strings.TrimSpace(parts[0])
		subparts := strings.Split(parts[1], " ")
		for j := 0; j < len(subparts); j++ {
			values := strings.Split(strings.TrimSpace(subparts[j]), "=")
			if len(values) == 2 {
				v, err := strconv.ParseInt(values[1], 10, 64)
				if err != nil {
					return err
				}
				fields[prefix+"_"+values[0]] = int64(v)
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
