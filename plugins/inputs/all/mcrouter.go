//go:build !custom || inputs || inputs.mcrouter

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/mcrouter"
)
