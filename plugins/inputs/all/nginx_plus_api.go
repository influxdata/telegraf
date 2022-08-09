//go:build all || inputs || inputs.nginx_plus_api

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/nginx_plus_api"
)
