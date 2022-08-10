//go:build !custom || inputs || inputs.sflow

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/sflow"
)
