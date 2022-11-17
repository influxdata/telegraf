//go:build !custom || inputs || inputs.azure_monitor

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/azure_monitor" // register plugin
