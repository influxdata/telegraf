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

var slabinfoPath = "/proc/slabinfo" // so we can override it in tests

type Slabinfo struct {
}

func (_ Slabinfo) Description() string {
	return "Provides metrics on Linux's slab allocator."
}

func (_ Slabinfo) SampleConfig() string {
	return ""
}

func (si Slabinfo) Gather(acc telegraf.Accumulator) error {
	f, err := os.Open(slabinfoPath)
	if err != nil {
		return fmt.Errorf("unable to open %s: %s", slabinfoPath, err)
	}
	defer f.Close()

	scnr := bufio.NewScanner(f)

	if !scnr.Scan() {
		return fmt.Errorf("error reading headers: %s", scnr.Err())
	}
	line := scnr.Text()
	if line[0:8] == "slabinfo" {
		// version line. ignore
		scnr.Scan()
		line = scnr.Text()
	}
	if line[0:2] != "# " {
		return fmt.Errorf("unexpected slabinfo format")
	}
	line = line[2:]
	hdrs := strings.Fields(line)
	for i, hdr := range hdrs {
		if hdr == "name" {
			hdrs[i] = "name"
		} else if hdr[0] == '<' {
			hdrs[i] = hdr[1 : len(hdr)-1]
		} else {
			hdrs[i] = ""
		}
	}

	for scnr.Scan() {
		tags := map[string]string{}
		fields := map[string]interface{}{}

		cols := strings.Fields(scnr.Text())
		for i, col := range cols {
			switch hdrs[i] {
			case "":
				continue
			case "name":
				tags["name"] = col
			default:
				v, err := strconv.ParseUint(col, 10, 64)
				if err != nil {
					return fmt.Errorf("unable to parse %s: %s", col, err)
				}
				fields[hdrs[i]] = v
			}
		}

		acc.AddFields("slabinfo", fields, tags)
	}

	return nil
}
