//go:build !custom || processors || processors.tag_limit

package all

import (
	_ "github.com/influxdata/telegraf/plugins/processors/tag_limit"
)
