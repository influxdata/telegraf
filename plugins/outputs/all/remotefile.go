//go:build !custom || outputs || outputs.remotefile

package all

import _ "github.com/influxdata/telegraf/plugins/outputs/remotefile" // register plugin
