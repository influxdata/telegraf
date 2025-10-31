//go:build !custom || inputs || inputs.nftables

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/nftables" // register plugin
