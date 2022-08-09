//go:build all || inputs || inputs.sqlserver

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/sqlserver"
)
