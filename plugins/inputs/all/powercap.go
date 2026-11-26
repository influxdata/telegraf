//go:build !custom || inputs || inputs.powercap

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/powercap" // register plugin
