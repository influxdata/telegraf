//go:build !custom || outputs || outputs.cratedb

package all

import _ "github.com/influxdata/telegraf/plugins/outputs/cratedb" // register plugin
