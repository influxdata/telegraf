//go:build !custom || outputs || outputs.discard

package all

import (
	_ "github.com/influxdata/telegraf/plugins/outputs/discard"
)
