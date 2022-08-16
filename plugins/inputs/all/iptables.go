//go:build (!custom || inputs || inputs.iptables) && linux

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/iptables" // register plugin
