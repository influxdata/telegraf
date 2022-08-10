//go:build !custom || outputs || outputs.cloudwatch_logs

package all

import (
	_ "github.com/influxdata/telegraf/plugins/outputs/cloudwatch_logs"
)
