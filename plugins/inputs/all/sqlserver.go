//go:build !custom || inputs || inputs.sqlserver

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/sqlserver"
)
