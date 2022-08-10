//go:build !custom || inputs || inputs.ping

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/ping"
)
