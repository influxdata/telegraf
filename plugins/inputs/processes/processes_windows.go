// +build windows

package processes

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

func init() {
	inputs.Add("processes", func() telegraf.Input {
		return &inputs.InvalidPlugin{
			InvalidReason: "Process input plugin not supported on windows",
			OrigDesc:      "Get the number of processes and group them by status",
			OrigSampleCfg: "",
		}
	})
}
