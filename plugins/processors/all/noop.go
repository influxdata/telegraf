//go:build !custom || processors || processors.noop

package all

import _ "github.com/influxdata/telegraf/plugins/processors/noop" // register plugin
