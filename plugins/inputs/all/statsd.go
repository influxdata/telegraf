//go:build !custom || inputs || inputs.statsd

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/statsd" // register plugin
