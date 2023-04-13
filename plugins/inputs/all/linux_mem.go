//go:build !custom || inputs || inputs.linux_mem

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/linux_mem" // register plugin
