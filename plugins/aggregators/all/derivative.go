//go:build !custom || aggregators || aggregators.derivative

package all

import _ "github.com/influxdata/telegraf/plugins/aggregators/derivative" // register plugin
