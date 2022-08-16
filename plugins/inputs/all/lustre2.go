//go:build (!custom || inputs || inputs.lustre2) && !windows

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/lustre2" // register plugin
