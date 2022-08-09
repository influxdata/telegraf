//go:build all || inputs || inputs.filestat

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/filestat"
)
