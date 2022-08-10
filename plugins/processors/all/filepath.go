//go:build !custom || processors || processors.filepath

package all

import (
	_ "github.com/influxdata/telegraf/plugins/processors/filepath"
)
