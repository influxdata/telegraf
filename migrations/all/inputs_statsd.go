//go:build !custom || (migrations && (inputs || inputs.statsd))

package all

import _ "github.com/influxdata/telegraf/migrations/inputs_statsd" // register migration
