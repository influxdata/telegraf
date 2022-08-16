//go:build (!custom || inputs || inputs.intel_powerstat) && linux

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/intel_powerstat" // register plugin
