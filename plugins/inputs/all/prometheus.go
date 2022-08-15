//go:build !custom || inputs || inputs.prometheus

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/prometheus" // register plugin
