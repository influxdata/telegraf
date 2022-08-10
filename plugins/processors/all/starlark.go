//go:build !custom || processors || processors.starlark

package all

import (
	_ "github.com/influxdata/telegraf/plugins/processors/starlark"
)
