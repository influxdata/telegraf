//go:build !custom || outputs || outputs.mqtt

package all

import _ "github.com/influxdata/telegraf/plugins/outputs/mqtt" // register plugin
