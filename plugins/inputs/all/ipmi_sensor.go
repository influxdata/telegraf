//go:build !custom || inputs || inputs.ipmi_sensor

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/ipmi_sensor"
)
