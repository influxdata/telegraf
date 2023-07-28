//go:build !custom || inputs || inputs.kernel_ksm

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/kernel_ksm" // register plugin
