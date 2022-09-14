//go:build !custom || inputs || inputs.fibaro

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/fibaro" // register plugin
