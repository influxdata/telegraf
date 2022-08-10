//go:build !custom || inputs || inputs.varnish

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/varnish"
)
