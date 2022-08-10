//go:build !custom || outputs || outputs.instrumental

package all

import (
	_ "github.com/influxdata/telegraf/plugins/outputs/instrumental"
)
