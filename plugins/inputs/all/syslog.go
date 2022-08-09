//go:build all || inputs || inputs.syslog

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/syslog"
)
