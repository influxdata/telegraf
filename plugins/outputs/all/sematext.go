//go:build !custom || outputs || outputs.sematext

package all

import _ "github.com/influxdata/telegraf/plugins/outputs/sematext" // register plugin
