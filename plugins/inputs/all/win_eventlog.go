//go:build all || inputs || inputs.win_eventlog

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/win_eventlog"
)
