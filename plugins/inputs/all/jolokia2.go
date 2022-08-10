//go:build !custom || inputs || inputs.jolokia2

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/jolokia2"
)
