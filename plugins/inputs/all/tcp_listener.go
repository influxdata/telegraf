//go:build all || inputs || inputs.tcp_listener

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/tcp_listener"
)
