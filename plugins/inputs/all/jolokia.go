//go:build all || inputs || inputs.jolokia

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/jolokia"
)
