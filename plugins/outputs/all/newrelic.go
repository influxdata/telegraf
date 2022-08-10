//go:build !custom || outputs || outputs.newrelic

package all

import (
	_ "github.com/influxdata/telegraf/plugins/outputs/newrelic"
)
