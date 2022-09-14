//go:build !custom || outputs || outputs.nsq

package all

import _ "github.com/influxdata/telegraf/plugins/outputs/nsq" // register plugin
