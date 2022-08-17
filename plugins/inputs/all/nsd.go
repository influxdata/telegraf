//go:build !custom || inputs || inputs.nsd

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/nsd" // register plugin
