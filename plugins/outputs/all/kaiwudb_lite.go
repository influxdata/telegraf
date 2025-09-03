//go:build !custom || outputs || outputs.kaiwudb_lite

package all

import _ "github.com/influxdata/telegraf/plugins/outputs/kaiwudb_lite" // register plugin
