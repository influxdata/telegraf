//go:build all || inputs || inputs.directory_monitor

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/directory_monitor"
)
