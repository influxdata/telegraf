//go:build !custom || outputs || outputs.sql

package all

import (
	_ "github.com/influxdata/telegraf/plugins/outputs/sql"
)
