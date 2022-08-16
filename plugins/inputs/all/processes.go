//go:build (!custom || inputs || inputs.processes) && !windows

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/processes" // register plugin
