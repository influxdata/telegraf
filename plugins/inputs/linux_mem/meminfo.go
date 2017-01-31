// +build linux

package linux_mem

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
)

var meminfoPath = "/proc/meminfo" // so we can override it in tests

type Meminfo struct {
}

func (_ Meminfo) Description() string {
	return "Provides detailed metrics about linux's memory usage."
}

func (_ Meminfo) SampleConfig() string {
	return ""
}

func (mi Meminfo) Gather(acc telegraf.Accumulator) error {
	f, err := os.Open(meminfoPath)
	if err != nil {
		return fmt.Errorf("unable to open %s: %s", meminfoPath, err)
	}
	defer f.Close()

	fields := map[string]interface{}{}
	scnr := bufio.NewScanner(f)
	for scnr.Scan() {
		spl := strings.Split(scnr.Text(), ":")

		k := strings.Replace(strings.Replace(spl[0], ")", "", 1), "(", "_", 1)

		v, err := strconv.ParseUint(strings.Trim(spl[1], " kB"), 10, 64)
		if err != nil {
			return fmt.Errorf("error parsing value from '%s': %s", spl[1], err)
		}
		fields[k] = v
	}
	if err := scnr.Err(); err != nil {
		return fmt.Errorf("error reading /proc/meminfo: %s", err)
	}
	acc.AddFields("mem", fields, nil)

	return nil
}
