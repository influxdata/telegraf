//go:build !custom || inputs || inputs.socket_listener

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/socket_listener" // register plugin
