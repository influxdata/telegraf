//go:build !custom || inputs || inputs.net

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/net"
)
