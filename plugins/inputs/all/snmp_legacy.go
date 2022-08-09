//go:build all || inputs || inputs.snmp_legacy

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/snmp_legacy"
)
