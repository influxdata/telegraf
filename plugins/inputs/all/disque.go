//go:build !custom || inputs || inputs.disque

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/disque"
)
