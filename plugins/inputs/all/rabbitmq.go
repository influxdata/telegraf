//go:build !custom || inputs || inputs.rabbitmq

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/rabbitmq" // register plugin
