//go:build !custom || parsers || parsers.graphite

package all

import (
	_ "github.com/influxdata/telegraf/plugins/parsers/graphite"
)
