//go:build all || inputs || inputs.conntrack

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/conntrack"
)
