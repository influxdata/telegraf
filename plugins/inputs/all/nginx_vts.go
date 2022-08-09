//go:build all || inputs || inputs.nginx_vts

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/nginx_vts"
)
