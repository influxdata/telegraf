//go:build !custom || inputs || inputs.win_wmi

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/win_wmi" // register plugin
