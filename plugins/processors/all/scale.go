//go:build !custom || processors || processors.scale

package all

import _ "github.com/influxdata/telegraf/plugins/processors/scale" // register plugin
