//go:build !custom || parsers || parsers.parquet

package all

import _ "github.com/influxdata/telegraf/plugins/parsers/parquet" // register plugin
