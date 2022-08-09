//go:build all || inputs || inputs.upsd

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/upsd"
)
