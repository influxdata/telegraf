//go:build !custom || inputs || inputs.promql

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/promql" // register plugin
