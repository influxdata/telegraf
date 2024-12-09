//go:build !custom || inputs || inputs.iptables

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/iptables" // register plugin
