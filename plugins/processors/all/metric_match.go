//go:build !custom || processors || processors.metric_match

package all

import _ "github.com/influxdata/telegraf/plugins/processors/metric_match" // register plugin
