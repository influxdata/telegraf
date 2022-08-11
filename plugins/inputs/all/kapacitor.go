//go:build !custom || inputs || inputs.kapacitor

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/kapacitor" // register plugin
