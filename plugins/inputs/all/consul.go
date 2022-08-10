//go:build !custom || inputs || inputs.consul

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/consul"
)
