//go:build (!custom || inputs || inputs.logparser) && !solaris

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/logparser" // register plugin
