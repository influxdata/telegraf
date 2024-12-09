//go:build !custom || inputs || inputs.intel_dlb

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/intel_dlb" // register plugin
