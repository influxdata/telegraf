//go:build !custom || outputs || outputs.cloud_pubsub

package all

import (
	_ "github.com/influxdata/telegraf/plugins/outputs/cloud_pubsub"
)
