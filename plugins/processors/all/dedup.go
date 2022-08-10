//go:build !custom || processors || processors.dedup

package all

import (
	_ "github.com/influxdata/telegraf/plugins/processors/dedup"
)
