//go:build !custom || inputs || inputs.kernel_vmstat

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/kernel_vmstat" // register plugin
