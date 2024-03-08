//go:build !custom || processors || processors.filter_ex

package all

import _ "github.com/influxdata/telegraf/plugins/processors/filter_ex" // register plugin
