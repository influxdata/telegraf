//go:build !custom || parsers || parsers.nagios

package all

import (
	_ "github.com/influxdata/telegraf/plugins/parsers/nagios"
)
