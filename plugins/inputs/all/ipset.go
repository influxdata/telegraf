//go:build !custom || inputs || inputs.ipset

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/ipset"
)
