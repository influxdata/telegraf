//go:build all || inputs || inputs.nginx

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/nginx"
)
