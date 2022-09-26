//go:build !custom || aggregators || aggregators.final

package all

import _ "github.com/influxdata/telegraf/plugins/aggregators/final" // register plugin
