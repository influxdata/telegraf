//go:build all || inputs || inputs.sql

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/sql"
)
