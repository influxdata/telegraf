//go:build !custom || processors || processors.defaults

package all

import (
	_ "github.com/influxdata/telegraf/plugins/processors/defaults"
)
