//go:build !custom || inputs || inputs.mongodb

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/mongodb"
)
