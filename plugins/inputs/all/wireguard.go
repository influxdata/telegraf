//go:build !custom || inputs || inputs.wireguard

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/wireguard" // register plugin
