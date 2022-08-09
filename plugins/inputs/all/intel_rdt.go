//go:build all || inputs || inputs.intel_rdt

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/intel_rdt"
)
