//go:build !custom || inputs || inputs.nginx

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/nginx" // register plugin
