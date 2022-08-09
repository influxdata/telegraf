//go:build all || inputs || inputs.wireguard

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/wireguard"
)
