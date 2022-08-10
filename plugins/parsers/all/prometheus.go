//go:build !custom || parsers || parsers.prometheus

package all

import (
	_ "github.com/influxdata/telegraf/plugins/parsers/prometheus"
)
