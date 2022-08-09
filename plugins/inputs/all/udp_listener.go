//go:build all || inputs || inputs.udp_listener

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/udp_listener"
)
