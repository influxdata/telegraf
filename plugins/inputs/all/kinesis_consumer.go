//go:build !custom || inputs || inputs.kinesis_consumer

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/kinesis_consumer"
)
