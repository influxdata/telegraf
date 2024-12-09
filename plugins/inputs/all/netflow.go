//go:build !custom || inputs || inputs.netflow

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/netflow" // register plugin
