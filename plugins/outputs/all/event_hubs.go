//go:build !custom || outputs || outputs.event_hubs

package all

import (
	_ "github.com/influxdata/telegraf/plugins/outputs/event_hubs"
)
