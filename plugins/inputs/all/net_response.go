//go:build !custom || inputs || inputs.net_response

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/net_response"
)
