//go:build !custom || processors || processors.strings

package all

import _ "github.com/influxdata/telegraf/plugins/processors/strings" // register plugin
