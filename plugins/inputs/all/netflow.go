//go:build !custom || inputs || inputs.netlow

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/netflow" // register plugin
