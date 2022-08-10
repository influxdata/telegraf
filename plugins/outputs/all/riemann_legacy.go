//go:build !custom || outputs || outputs.riemann_legacy

package all

import (
	_ "github.com/influxdata/telegraf/plugins/outputs/riemann_legacy"
)
