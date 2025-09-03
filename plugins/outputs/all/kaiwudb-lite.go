//go:build !custom || outputs || outputs.kaiwudb

package all

import _ "github.com/influxdata/telegraf/plugins/outputs/kaiwudb-lite" // register plugin
