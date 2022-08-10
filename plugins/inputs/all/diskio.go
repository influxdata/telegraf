//go:build !custom || inputs || inputs.diskio || core

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/diskio"
)
