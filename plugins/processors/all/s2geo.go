//go:build !custom || processors || processors.s2geo

package all

import (
	_ "github.com/influxdata/telegraf/plugins/processors/s2geo"
)
