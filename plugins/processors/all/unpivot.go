//go:build !custom || processors || processors.unpivot

package all

import (
	_ "github.com/influxdata/telegraf/plugins/processors/unpivot"
)
