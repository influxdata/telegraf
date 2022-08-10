//go:build !custom || outputs || outputs.signalfx

package all

import (
	_ "github.com/influxdata/telegraf/plugins/outputs/signalfx"
)
