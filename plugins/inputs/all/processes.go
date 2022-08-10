//go:build !custom || inputs || inputs.processes || core

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/processes"
)
