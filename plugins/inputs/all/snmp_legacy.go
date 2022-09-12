//go:build !custom || inputs || inputs.snmp_legacy

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/snmp_legacy" // register plugin
