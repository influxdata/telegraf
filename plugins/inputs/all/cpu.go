//go:build !custom || inputs || inputs.cpu

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/cpu" // register plugin
