//go:build !custom || inputs || inputs.cisco_telemetry_mdt

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/cisco_telemetry_mdt" // register plugin
