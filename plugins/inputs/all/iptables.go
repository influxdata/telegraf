//go:build all || inputs || inputs.iptables

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/iptables"
)
