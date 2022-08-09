//go:build all || inputs || inputs.win_services

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/win_services"
)
