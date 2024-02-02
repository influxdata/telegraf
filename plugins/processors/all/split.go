//go:build !custom || processors || processors.split

package all

import _ "github.com/influxdata/telegraf/plugins/processors/split" // register plugin
