//go:build all || inputs || inputs.nats_consumer

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/nats_consumer"
)
