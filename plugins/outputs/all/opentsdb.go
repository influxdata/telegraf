//go:build !custom || outputs || outputs.opentsdb

package all

import (
	_ "github.com/influxdata/telegraf/plugins/outputs/opentsdb"
)
