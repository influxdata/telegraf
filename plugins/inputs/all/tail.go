//go:build !custom || inputs || inputs.tail

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/tail"
)
