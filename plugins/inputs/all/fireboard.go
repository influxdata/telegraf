//go:build !custom || inputs || inputs.fireboard

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/fireboard"
)
