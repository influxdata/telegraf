//go:build all || inputs || inputs.influxdb_listener

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/influxdb_listener"
)
