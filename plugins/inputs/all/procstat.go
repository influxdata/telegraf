//go:build !custom || inputs || inputs.procstat

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/procstat" // register plugin
