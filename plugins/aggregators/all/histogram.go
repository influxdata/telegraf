//go:build !custom || aggregators || aggregators.histogram

package all

import _ "github.com/influxdata/telegraf/plugins/aggregators/histogram" // register plugin
