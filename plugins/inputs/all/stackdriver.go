//go:build all || inputs || inputs.stackdriver

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/stackdriver"
)
