//go:build !custom || outputs || outputs.stackdriver

package all

import (
	_ "github.com/influxdata/telegraf/plugins/outputs/stackdriver"
)
