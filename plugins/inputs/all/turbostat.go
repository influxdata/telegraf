//go:build !custom || inputs || inputs.turbostat

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/turbostat" // register plugin
