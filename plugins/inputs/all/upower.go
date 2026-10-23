//go:build !custom || inputs || inputs.upower

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/upower" // register plugin
