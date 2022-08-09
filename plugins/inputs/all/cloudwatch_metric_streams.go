//go:build all || inputs || inputs.cloudwatch_metric_streams

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/cloudwatch_metric_streams"
)
