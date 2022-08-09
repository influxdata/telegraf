//go:build all || inputs || inputs.knx_listener

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/knx_listener"
)
