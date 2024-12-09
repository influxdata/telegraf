//go:build !custom || aggregators || aggregators.valuecounter

package all

import _ "github.com/influxdata/telegraf/plugins/aggregators/valuecounter" // register plugin
