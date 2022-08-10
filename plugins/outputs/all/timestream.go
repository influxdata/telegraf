//go:build !custom || outputs || outputs.timestream

package all

import (
	_ "github.com/influxdata/telegraf/plugins/outputs/timestream"
)
