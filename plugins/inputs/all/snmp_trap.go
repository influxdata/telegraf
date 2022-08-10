//go:build !custom || inputs || inputs.snmp_trap

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/snmp_trap"
)
