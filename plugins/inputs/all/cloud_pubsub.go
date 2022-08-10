//go:build !custom || inputs || inputs.cloud_pubsub

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/cloud_pubsub"
)
