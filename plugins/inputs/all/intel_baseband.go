//go:build !custom || inputs || inputs.intel_baseband

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/intel_baseband" // register plugin
