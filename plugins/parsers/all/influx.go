//go:build !custom || parsers || parsers.influx

package all

import (
	_ "github.com/influxdata/telegraf/plugins/parsers/influx"                 // register plugin
	_ "github.com/influxdata/telegraf/plugins/parsers/influx/influx_upstream" // register plugin
)
