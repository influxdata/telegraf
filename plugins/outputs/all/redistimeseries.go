//go:build !custom || outputs || outputs.redistimeseries

package all

import (
	_ "github.com/influxdata/telegraf/plugins/outputs/redistimeseries"
)
