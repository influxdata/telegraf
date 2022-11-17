//go:build !custom || processors || processors.execd

package all

import _ "github.com/influxdata/telegraf/plugins/processors/execd" // register plugin
