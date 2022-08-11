//go:build !custom || inputs || inputs.conntrack

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/conntrack" // register plugin
