//go:build !custom || outputs || outputs.zerobus

package all

import _ "github.com/influxdata/telegraf/plugins/outputs/zerobus" // register plugin
