//go:build !custom || inputs || inputs.puppetagent

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/puppetagent"
)
