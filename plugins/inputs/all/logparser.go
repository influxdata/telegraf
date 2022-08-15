//go:build !custom || inputs || inputs.logparser

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/logparser" // register plugin
