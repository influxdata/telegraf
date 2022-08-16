//go:build (!custom || inputs || inputs.sensors) && linux

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/sensors" // register plugin
