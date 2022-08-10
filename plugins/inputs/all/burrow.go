//go:build !custom || inputs || inputs.burrow

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/burrow"
)
