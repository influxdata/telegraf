//go:build !custom || inputs || inputs.filecount

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/filecount"
)
