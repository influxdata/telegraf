//go:build !custom || inputs || inputs.interrupts

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/interrupts"
)
