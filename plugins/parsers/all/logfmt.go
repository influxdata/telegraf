//go:build !custom || parsers || parsers.logfmt

package all

import (
	_ "github.com/influxdata/telegraf/plugins/parsers/logfmt"
)
