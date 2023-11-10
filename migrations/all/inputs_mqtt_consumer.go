//go:build !custom || (migrations && (inputs || inputs.mqtt_consumer))

package all

import _ "github.com/influxdata/telegraf/migrations/inputs_mqtt_consumer" // register migration
