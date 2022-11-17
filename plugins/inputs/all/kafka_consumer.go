//go:build !custom || inputs || inputs.kafka_consumer

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/kafka_consumer" // register plugin
