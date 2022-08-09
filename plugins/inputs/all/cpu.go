//go:build all || inputs || inputs.cpu || core

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/cpu"
)
