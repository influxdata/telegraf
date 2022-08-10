//go:build !custom || inputs || inputs.trig

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/trig"
)
