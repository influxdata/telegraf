//go:build !custom || inputs || inputs.activemq

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/traceroute" // register plugin
