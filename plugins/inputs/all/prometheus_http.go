//go:build !custom || inputs || inputs.prometheus_http

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/prometheus_http" // register plugin
