//go:build !custom || parsers || parsers.grok

package all

import (
	_ "github.com/influxdata/telegraf/plugins/parsers/grok"
)
