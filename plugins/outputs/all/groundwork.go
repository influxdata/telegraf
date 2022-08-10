//go:build !custom || outputs || outputs.groundwork

package all

import (
	_ "github.com/influxdata/telegraf/plugins/outputs/groundwork"
)
