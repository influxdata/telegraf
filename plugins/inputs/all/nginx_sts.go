//go:build !custom || inputs || inputs.nginx_sts

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/nginx_sts" // register plugin
