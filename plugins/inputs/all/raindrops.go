//go:build !custom || inputs || inputs.raindrops

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/raindrops" // register plugin
