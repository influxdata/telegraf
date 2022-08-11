//go:build !custom || outputs || outputs.influxdb

package all

import _ "github.com/influxdata/telegraf/plugins/outputs/influxdb" // register plugin
