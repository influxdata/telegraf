//go:build all || inputs || inputs.kernel || core

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/kernel"
)
