//go:build all || inputs || inputs.eventhub_consumer

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/eventhub_consumer"
)
