//go:build all || inputs || inputs.monit

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/monit"
)
