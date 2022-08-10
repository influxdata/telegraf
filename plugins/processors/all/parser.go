//go:build !custom || processors || processors.parser

package all

import (
	_ "github.com/influxdata/telegraf/plugins/processors/parser"
)
