//go:build !custom || processors || processors.template

package all

import (
	_ "github.com/influxdata/telegraf/plugins/processors/template"
)
