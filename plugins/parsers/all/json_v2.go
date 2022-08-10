//go:build !custom || parsers || parsers.json_v2

package all

import (
	_ "github.com/influxdata/telegraf/plugins/parsers/json_v2"
)
