//go:build !custom || (migrations && (outputs || outputs.influxdb))

package all

import _ "github.com/influxdata/telegraf/migrations/outputs_influxdb" // register migration
