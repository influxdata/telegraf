//go:build (!custom || inputs || inputs.tail) && !solaris

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/tail" // register plugin
