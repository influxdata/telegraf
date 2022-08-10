//go:build !custom || outputs || outputs.amqp

package all

import (
	_ "github.com/influxdata/telegraf/plugins/outputs/amqp"
)
