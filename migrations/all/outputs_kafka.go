//go:build !custom || (migrations && (outputs || outputs.kafka))

package all

import _ "github.com/influxdata/telegraf/migrations/outputs_kafka" // register migration
