//go:build !custom || processors || processors.rename

package all

import (
	_ "github.com/influxdata/telegraf/plugins/processors/rename"
)
