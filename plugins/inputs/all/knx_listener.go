//go:build !custom || inputs || inputs.knx_listener

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/knx_listener" // register plugin
