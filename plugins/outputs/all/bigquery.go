//go:build !custom || outputs || outputs.bigquery

package all

import _ "github.com/influxdata/telegraf/plugins/outputs/bigquery" // register plugin
