//go:build (!custom || inputs || inputs.nats) && (!freebsd || cgo)

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/nats" // register plugin
