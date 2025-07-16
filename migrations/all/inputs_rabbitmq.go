//go:build !custom || (migrations && (inputs || inputs.rabbitmq))

package all

import _ "github.com/influxdata/telegraf/migrations/inputs_rabbitmq" // register migration
