//go:build all || inputs || inputs.socketstat

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/socketstat"
)
