package outputs

import "github.com/influxdata/telegraf"

var Deprecations = map[string]telegraf.DeprecationInfo{
	"riemann_legacy": {
		Since:  "1.3",
		Notice: "use 'outputs.riemann' instead (see https://github.com/influxdata/telegraf/issues/1878)",
	},
}
