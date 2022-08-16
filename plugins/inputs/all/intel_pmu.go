//go:build (!custom || inputs || inputs.intel_pmu) && linux && amd64

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/intel_pmu" // register plugin
