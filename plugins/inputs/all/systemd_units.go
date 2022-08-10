//go:build !custom || inputs || inputs.systemd_units

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/systemd_units"
)
