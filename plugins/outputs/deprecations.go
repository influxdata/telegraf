package outputs

import "github.com/influxdata/telegraf"

// Deprecations lists the deprecated plugins
var Deprecations = map[string]telegraf.DeprecationInfo{
	"riemann_legacy": {
		Since:     "1.3.0",
		RemovalIn: "1.30.0",
		Notice:    "use 'outputs.riemann' instead (see https://github.com/influxdata/telegraf/issues/1878)",
	},
	"amon": {
		Since:     "1.37.0",
		RemovalIn: "1.40.0",
		Notice:    "service doesn't exist anymore and platform code is unmaintained",
	},
}
