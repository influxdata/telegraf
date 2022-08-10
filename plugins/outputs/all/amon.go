//go:build !custom || outputs || outputs.amon

package all

import (
	_ "github.com/influxdata/telegraf/plugins/outputs/amon"
)
