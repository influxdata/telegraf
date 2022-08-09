//go:build all || inputs || inputs.intel_pmu

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/intel_pmu"
)
