//go:build (!custom || inputs || inputs.ipvs) && linux

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/ipvs" // register plugin
