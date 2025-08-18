//go:build !custom || (migrations && (outputs || outputs.mqtt))

package all

import _ "github.com/influxdata/telegraf/migrations/outputs_mqtt" // register migration
