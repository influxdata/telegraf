//go:build !custom || inputs || inputs.execd

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/execd" // register plugin
