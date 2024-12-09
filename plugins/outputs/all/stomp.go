//go:build !custom || outputs || outputs.stomp

package all

import _ "github.com/influxdata/telegraf/plugins/outputs/stomp" // register plugin
