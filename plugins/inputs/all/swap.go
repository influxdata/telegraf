//go:build !custom || inputs || inputs.swap || core

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/swap"
)
