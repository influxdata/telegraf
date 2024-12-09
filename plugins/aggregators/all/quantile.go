//go:build !custom || aggregators || aggregators.quantile

package all

import _ "github.com/influxdata/telegraf/plugins/aggregators/quantile" // register plugin
