//go:build all || inputs || inputs.opentelemetry

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/opentelemetry"
)
