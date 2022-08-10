//go:build !custom || inputs || inputs.disk || core

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/disk"
)
