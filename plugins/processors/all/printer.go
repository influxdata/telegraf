//go:build !custom || processors || processors.printer

package all

import (
	_ "github.com/influxdata/telegraf/plugins/processors/printer"
)
