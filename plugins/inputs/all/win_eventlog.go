//go:build (!custom || inputs || inputs.win_eventlog) && windows

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/win_eventlog" // register plugin
