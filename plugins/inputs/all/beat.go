//go:build all || inputs || inputs.beat

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/beat"
)
