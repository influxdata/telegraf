//go:build !custom || aggregators || aggregators.starlark

package all

import _ "github.com/influxdata/telegraf/plugins/aggregators/starlark" // register plugin
