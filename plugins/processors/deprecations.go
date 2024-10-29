package processors

import "github.com/influxdata/telegraf"

// Deprecations lists the deprecated plugins
var Deprecations = make(map[string]telegraf.DeprecationInfo)
