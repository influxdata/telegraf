//go:build all || inputs || inputs.mem || core

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/mem"
)
