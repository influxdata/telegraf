//go:build !custom || inputs || inputs.minecraft

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/minecraft"
)
