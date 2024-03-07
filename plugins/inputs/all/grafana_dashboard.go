//go:build !custom || inputs || inputs.grafana_dashboard

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/grafana_dashboard" // register plugin
