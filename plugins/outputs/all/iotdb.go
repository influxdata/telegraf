//go:build !custom || outputs || outputs.iotdb

package all

import _ "github.com/influxdata/telegraf/plugins/outputs/iotdb" // register plugin
