//go:build !custom || inputs || inputs.rapl

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/rapl" // register plugin
