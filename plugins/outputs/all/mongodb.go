//go:build !custom || outputs || outputs.mongodb

package all

import (
	_ "github.com/influxdata/telegraf/plugins/outputs/mongodb"
)
