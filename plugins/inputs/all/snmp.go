//go:build all || inputs || inputs.snmp

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/snmp"
)
