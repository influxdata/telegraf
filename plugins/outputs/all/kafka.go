//go:build !custom || outputs || outputs.kafka

package all

import (
	_ "github.com/influxdata/telegraf/plugins/outputs/kafka"
)
