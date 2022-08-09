//go:build all || inputs || inputs.win_perf_counters

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/win_perf_counters"
)
