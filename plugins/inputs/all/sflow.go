//go:build all || inputs || inputs.sflow

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/sflow"
)
