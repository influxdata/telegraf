//go:build !custom || inputs || inputs.gnmi_listener

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/gnmi_listener" // register plugin
