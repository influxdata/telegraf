//go:build all || inputs || inputs.twemproxy

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/twemproxy"
)
