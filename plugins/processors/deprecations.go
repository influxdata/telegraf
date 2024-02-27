package processors

import "github.com/influxdata/telegraf"

// Deprecations lists the deprecated plugins
var Deprecations = map[string]telegraf.DeprecationInfo{
	"ifname": {
		Since:  "1.30.0",
		Notice: "use 'processors.snmp_lookup' instead",
	},
}
