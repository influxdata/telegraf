//go:build !custom || outputs || outputs.opentelemetry

package all

import (
	_ "github.com/influxdata/telegraf/plugins/outputs/opentelemetry"
)
