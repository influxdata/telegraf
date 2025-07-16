//go:build !custom || (migrations && (inputs || inputs.influxdb_listener))

package all

import _ "github.com/influxdata/telegraf/migrations/inputs_influxdb_listener" // register migration
