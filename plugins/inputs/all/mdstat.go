//go:build !custom || inputs || inputs.mdstat

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/mdstat" // register plugin
