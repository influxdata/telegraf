//go:build all || inputs || inputs.sysstat

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/sysstat"
)
