//go:build !custom || processors || processors.pivot

package all

import (
	_ "github.com/influxdata/telegraf/plugins/processors/pivot"
)
