//go:build !custom || inputs || inputs.aiven_procstat

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/aiven-procstat" // register plugin
