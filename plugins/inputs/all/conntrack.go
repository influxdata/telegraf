//go:build (!custom || inputs || inputs.conntrack) && linux

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/conntrack" // register plugin
