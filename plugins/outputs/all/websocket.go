//go:build !custom || outputs || outputs.websocket

package all

import _ "github.com/influxdata/telegraf/plugins/outputs/websocket" // register plugin
