//go:build !custom || inputs || inputs.pgbouncer

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/pgbouncer"
)
