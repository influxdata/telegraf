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

var buddyinfoPath = "/proc/buddyinfo" // so we can override it in tests

type Buddyinfo struct {
	UseOrder bool
}

func (_ Buddyinfo) Description() string {
	return "Provides metrics on linux memory page availability."
}

func (_ Buddyinfo) SampleConfig() string {
	return `
  ## Report fields in exponential order instead of KB units.
  # use_order = true
`
}

func (bi Buddyinfo) Gather(acc telegraf.Accumulator) error {
	f, err := os.Open(buddyinfoPath)
	if err != nil {
		return fmt.Errorf("unable to open %s: %s", buddyinfoPath, err)
	}
	defer f.Close()

	var pageSizeKB int
	if !bi.UseOrder {
		pageSizeKB = os.Getpagesize() / 1024
	}

	scnr := bufio.NewScanner(f)
	for scnr.Scan() {
		tags := map[string]string{}
		fields := map[string]interface{}{}

		cols := strings.Fields(scnr.Text())
		if len(cols) != 15 {
			return fmt.Errorf("unexpected number of fields in %s. Expected 15, found %d", buddyinfoPath, len(cols))
		}
		if cols[1][len(cols[1])-1] != ',' {
			return fmt.Errorf("unexpected format in %s", buddyinfoPath)
		}

		tags["node"] = cols[1][:len(cols[1])-1]
		tags["zone"] = cols[3]

		for i := uint(0); i < 11; i++ {
			pages, err := strconv.ParseUint(cols[i+4], 10, 64)
			if err != nil {
				return fmt.Errorf("unable to parse %s: %s", cols[i+4], err)
			}

			var field string
			if !bi.UseOrder {
				field = fmt.Sprintf("%dk", pageSizeKB<<i)
			} else {
				field = fmt.Sprintf("order-%d", i)
			}
			fields[field] = pages

			acc.AddFields("buddyinfo", fields, tags)
		}
	}

	return nil
}
