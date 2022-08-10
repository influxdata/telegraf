//go:build !custom || parsers || parsers.wavefront

package all

import (
	_ "github.com/influxdata/telegraf/plugins/parsers/wavefront"
)
