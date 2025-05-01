//go:build !custom || (migrations && (outputs || outputs.amqp))

package all

import _ "github.com/influxdata/telegraf/migrations/outputs_amqp" // register migration
