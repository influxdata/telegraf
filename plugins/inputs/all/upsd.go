//go:build !custom || inputs || inputs.upsd

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/upsd" // register plugin
