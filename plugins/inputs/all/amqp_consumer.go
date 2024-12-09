//go:build !custom || inputs || inputs.amqp_consumer

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/amqp_consumer" // register plugin
