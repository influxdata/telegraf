//go:build all || inputs || inputs.webhooks

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/webhooks"
)
