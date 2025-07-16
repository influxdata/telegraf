//go:build !custom || (migrations && (inputs || inputs.cisco_telemetry_gnmi))

package all

import _ "github.com/influxdata/telegraf/migrations/inputs_cisco_telemetry_gnmi" // register migration
