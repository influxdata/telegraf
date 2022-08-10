//go:build !custom || outputs || outputs.logzio

package all

import (
	_ "github.com/influxdata/telegraf/plugins/outputs/logzio"
)
