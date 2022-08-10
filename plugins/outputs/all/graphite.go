//go:build !custom || outputs || outputs.graphite

package all

import (
	_ "github.com/influxdata/telegraf/plugins/outputs/graphite"
)
