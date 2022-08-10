//go:build !custom || inputs || inputs.zipkin

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/zipkin"
)
