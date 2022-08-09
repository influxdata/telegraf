//go:build all || inputs || inputs.chrony

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/chrony"
)
