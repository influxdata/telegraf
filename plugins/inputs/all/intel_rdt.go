//go:build (!custom || inputs || inputs.intel_rdt) && !windows

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/intel_rdt" // register plugin
