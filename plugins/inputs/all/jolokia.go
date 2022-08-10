//go:build !custom || inputs || inputs.jolokia

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/jolokia"
)
