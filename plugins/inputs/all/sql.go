//go:build !custom || inputs || inputs.sql

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/sql"
)
