//go:build !custom || inputs || inputs.kernel

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/kernel" // register plugin
