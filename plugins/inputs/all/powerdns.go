//go:build !custom || inputs || inputs.powerdns

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/powerdns"
)
