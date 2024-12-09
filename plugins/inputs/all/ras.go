//go:build !custom || inputs || inputs.ras

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/ras" // register plugin
