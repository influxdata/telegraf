//go:build !custom || (migrations && (inputs || inputs.kafka_consumer_legacy))

package all

import _ "github.com/influxdata/telegraf/migrations/inputs_kafka_consumer_legacy" // register migration
