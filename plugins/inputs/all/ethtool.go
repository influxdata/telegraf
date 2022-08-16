//go:build (!custom || inputs || inputs.ethtool) && linux

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/ethtool" // register plugin
