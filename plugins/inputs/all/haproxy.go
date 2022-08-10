//go:build !custom || inputs || inputs.haproxy

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/haproxy"
)
