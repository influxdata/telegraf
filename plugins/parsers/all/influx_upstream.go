//go:build !custom || parsers || parsers.influx_upstream

package all

import (
	_ "github.com/influxdata/telegraf/plugins/parsers/influx_upstream"
)
