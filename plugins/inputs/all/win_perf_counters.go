//go:build (!custom || inputs || inputs.win_perf_counters) && windows

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/win_perf_counters" // register plugin
