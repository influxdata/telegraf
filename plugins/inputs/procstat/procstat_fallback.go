// +build !linux

package procstat

import (
	"fmt"

	"github.com/influxdata/telegraf"
)

func addConnectionStats(pidConnections []ConnInfo, fields map[string]interface{}, prefix string) {
}

func addConnectionEnpoints(acc telegraf.Accumulator, proc Process, netInfo NetworkInfo) error {
	return fmt.Errorf("platform not supported")
}
