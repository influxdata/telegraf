//go:build !custom || aggregators || aggregators.basicstats

package all

import _ "github.com/influxdata/telegraf/plugins/aggregators/basicstats" // register plugin
