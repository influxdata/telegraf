//go:build !custom || processors || processors.regex

package all

import (
	_ "github.com/influxdata/telegraf/plugins/processors/regex"
)
