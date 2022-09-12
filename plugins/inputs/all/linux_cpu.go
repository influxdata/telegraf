//go:build !custom || inputs || inputs.linux_cpu

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/linux_cpu" // register plugin
