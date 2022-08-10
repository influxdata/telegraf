//go:build !custom || processors || processors.port_name

package all

import (
	_ "github.com/influxdata/telegraf/plugins/processors/port_name"
)
