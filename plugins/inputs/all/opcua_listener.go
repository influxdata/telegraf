//go:build !custom || inputs || inputs.opcua_listener

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/opcua_listener" // register plugin
