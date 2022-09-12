//go:build !custom || outputs || outputs.dynatrace

package all

import _ "github.com/influxdata/telegraf/plugins/outputs/dynatrace" // register plugin
