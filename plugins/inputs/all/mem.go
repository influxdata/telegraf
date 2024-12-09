//go:build !custom || inputs || inputs.mem

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/mem" // register plugin
