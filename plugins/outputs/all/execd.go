//go:build !custom || outputs || outputs.execd

package all

import _ "github.com/influxdata/telegraf/plugins/outputs/execd" // register plugin
