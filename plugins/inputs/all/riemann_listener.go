//go:build all || inputs || inputs.riemann_listener

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/riemann_listener"
)
