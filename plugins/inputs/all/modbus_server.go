//go:build !custom || inputs || inputs.modbus_server

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/modbus_server" // register plugin
