//go:build !custom || inputs || inputs.logql

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/logql" // register plugin
