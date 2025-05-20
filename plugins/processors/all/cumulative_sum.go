//go:build !custom || processors || processors.cumulative_sum

package all

import _ "github.com/influxdata/telegraf/plugins/processors/cumulative_sum" // register plugin
