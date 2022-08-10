//go:build !custom || inputs || inputs.nsq_consumer

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/nsq_consumer"
)
