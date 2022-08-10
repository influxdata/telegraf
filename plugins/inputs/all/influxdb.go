//go:build !custom || inputs || inputs.influxdb

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/influxdb"
)
