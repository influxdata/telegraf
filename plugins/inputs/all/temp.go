//go:build !custom || inputs || inputs.temp

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/temp" // register plugin
