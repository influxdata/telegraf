//go:build all || inputs || inputs.modbus

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/modbus"
)
