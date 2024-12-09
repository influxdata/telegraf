//go:build !custom || processors || processors.timestamp

package all

import _ "github.com/influxdata/telegraf/plugins/processors/timestamp" // register plugin
