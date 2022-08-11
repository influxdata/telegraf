//go:build !custom || inputs || inputs.netstat

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/netstat" // register plugin
