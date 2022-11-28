//go:build !custom || inputs || inputs.pf

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/pf" // register plugin
