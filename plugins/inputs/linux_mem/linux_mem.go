// +build linux

package linux_mem

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

func init() {
	inputs.Add("linux_meminfo", func() telegraf.Input {
		return Meminfo{}
	})
	inputs.Add("linux_buddyinfo", func() telegraf.Input {
		return Buddyinfo{}
	})
	inputs.Add("linux_slabinfo", func() telegraf.Input {
		return Slabinfo{}
	})
}
