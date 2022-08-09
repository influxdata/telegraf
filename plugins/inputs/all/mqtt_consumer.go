//go:build all || inputs || inputs.mqtt_consumer

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/mqtt_consumer"
)
