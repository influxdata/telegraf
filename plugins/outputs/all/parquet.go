//go:build !custom || outputs || outputs.parquet

package all

import _ "github.com/influxdata/telegraf/plugins/outputs/parquet" // register plugin
