//go:build !custom || outputs || outputs.modbus_server

package all

import _ "github.com/influxdata/telegraf/plugins/outputs/modbus_server" // register plugin
