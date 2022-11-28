//go:build !custom || outputs || outputs.nats

package all

import _ "github.com/influxdata/telegraf/plugins/outputs/nats" // register plugin
