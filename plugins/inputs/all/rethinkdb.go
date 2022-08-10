//go:build !custom || inputs || inputs.rethinkdb

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/rethinkdb"
)
