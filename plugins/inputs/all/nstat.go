//go:build !custom || inputs || inputs.nstat

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/nstat"
)
