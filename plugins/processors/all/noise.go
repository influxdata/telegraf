//go:build !custom || processors || processors.noise

package all

import (
	_ "github.com/influxdata/telegraf/plugins/processors/noise"
)
