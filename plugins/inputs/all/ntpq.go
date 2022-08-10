//go:build !custom || inputs || inputs.ntpq

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/ntpq"
)
