//go:build !custom || processors || processors.filter

package all

import _ "github.com/influxdata/telegraf/plugins/processors/filter" // register plugin
