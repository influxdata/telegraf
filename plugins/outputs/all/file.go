//go:build !custom || outputs || outputs.file

package all

import (
	_ "github.com/influxdata/telegraf/plugins/outputs/file"
)
