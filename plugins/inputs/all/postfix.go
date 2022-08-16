//go:build (!custom || inputs || inputs.postfix) && !windows

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/postfix" // register plugin
