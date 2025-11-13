//go:build !custom || outputs || outputs.heartbeat

package all

import _ "github.com/influxdata/telegraf/plugins/outputs/heartbeat" // register plugin
