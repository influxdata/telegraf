//go:build !custom || outputs || outputs.loki

package all

import (
	_ "github.com/influxdata/telegraf/plugins/outputs/loki"
)
