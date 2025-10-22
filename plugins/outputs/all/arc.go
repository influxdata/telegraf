//go:build !custom || outputs || outputs.arc

package all

import _ "github.com/influxdata/telegraf/plugins/outputs/arc" // register plugin
