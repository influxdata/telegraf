//go:build !custom || inputs || inputs.socketstat

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/socketstat"
)
