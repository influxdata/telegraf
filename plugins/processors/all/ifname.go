//go:build !custom || processors || processors.ifname

package all

import (
	_ "github.com/influxdata/telegraf/plugins/processors/ifname"
)
