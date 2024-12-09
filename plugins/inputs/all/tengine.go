//go:build !custom || inputs || inputs.tengine

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/tengine" // register plugin
