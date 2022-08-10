//go:build !custom || outputs || outputs.warp10

package all

import (
	_ "github.com/influxdata/telegraf/plugins/outputs/warp10"
)
