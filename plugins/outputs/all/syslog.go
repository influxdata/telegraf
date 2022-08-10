//go:build !custom || outputs || outputs.syslog

package all

import (
	_ "github.com/influxdata/telegraf/plugins/outputs/syslog"
)
