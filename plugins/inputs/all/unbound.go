//go:build !custom || inputs || inputs.unbound

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/unbound" // register plugin
