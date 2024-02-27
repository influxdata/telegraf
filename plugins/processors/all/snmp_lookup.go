//go:build !custom || processors || processors.snmp_lookup

package all

import _ "github.com/influxdata/telegraf/plugins/processors/snmp_lookup" // register plugin
