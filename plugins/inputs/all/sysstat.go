//go:build !custom || inputs || inputs.sysstat

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/sysstat" // register plugin
