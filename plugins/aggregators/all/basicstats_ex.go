//go:build !custom || aggregators || aggregators.basicstats_ex

package all

import _ "github.com/influxdata/telegraf/plugins/aggregators/basicstats_ex" // register plugin
