//go:build (!custom || inputs || inputs.sysstat) && linux

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/sysstat" // register plugin
