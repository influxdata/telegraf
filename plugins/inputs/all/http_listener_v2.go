//go:build all || inputs || inputs.http_listener_v2

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/http_listener_v2"
)
