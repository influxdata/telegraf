//go:build !custom || inputs || inputs.file

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/file"
)
