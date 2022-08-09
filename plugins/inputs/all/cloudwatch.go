//go:build all || inputs || inputs.cloudwatch

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/cloudwatch"
)
