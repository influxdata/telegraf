//go:build all || inputs || inputs.graylog

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/graylog"
)
