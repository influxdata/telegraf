//go:build !custom || inputs || inputs.snmp

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/snmp" // register plugin
