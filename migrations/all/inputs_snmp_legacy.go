//go:build !custom || (migrations && (inputs || inputs.snmp_legacy))

package all

import _ "github.com/influxdata/telegraf/migrations/inputs_snmp_legacy" // register migration
