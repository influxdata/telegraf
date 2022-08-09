//go:build all || inputs || inputs.system || core

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/system"
)
