//go:build !custom || inputs || inputs.nginx_upstream_check

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/nginx_upstream_check"
)
