//go:build (!custom || inputs || inputs.mdstat) && linux

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/mdstat" // register plugin
