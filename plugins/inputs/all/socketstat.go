//go:build (!custom || inputs || inputs.socketstat) && !windows

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/socketstat" // register plugin
