//go:build !custom || inputs || inputs.jti_openconfig_telemetry

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/jti_openconfig_telemetry"
)
