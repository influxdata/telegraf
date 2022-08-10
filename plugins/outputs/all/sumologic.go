//go:build !custom || outputs || outputs.sumologic

package all

import (
	_ "github.com/influxdata/telegraf/plugins/outputs/sumologic"
)
