//go:build (!custom || inputs || inputs.synproxy) && linux

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/synproxy" // register plugin
