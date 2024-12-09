//go:build !custom || inputs || inputs.nats

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/nats" // register plugin
