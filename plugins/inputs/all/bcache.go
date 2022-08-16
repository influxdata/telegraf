//go:build (!custom || inputs || inputs.bcache) && !windows

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/bcache" // register plugin
