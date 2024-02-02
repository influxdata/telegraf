//go:build !custom || inputs || inputs.intel_pmt

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/intel_pmt" // register plugin
