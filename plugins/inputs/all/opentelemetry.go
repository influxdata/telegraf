//go:build !custom || inputs || inputs.opentelemetry

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/opentelemetry" // register plugin
