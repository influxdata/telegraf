//go:build !custom || processors || processors.enum

package all

import (
	_ "github.com/influxdata/telegraf/plugins/processors/enum"
)
