//go:build !custom || processors || processors.traffic_shaper

package all

import _ "github.com/influxdata/telegraf/plugins/processors/traffic_shaper" // register plugin
