//go:build all || inputs || inputs.synproxy

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/synproxy"
)
