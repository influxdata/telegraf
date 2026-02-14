//go:build !custom || inputs || inputs.sip

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/sip" // register plugin
