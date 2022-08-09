//go:build all || inputs || inputs.intel_powerstat

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/intel_powerstat"
)
