//go:build !custom || outputs || outputs.cloudwatch

package all

import (
	_ "github.com/influxdata/telegraf/plugins/outputs/cloudwatch"
)
