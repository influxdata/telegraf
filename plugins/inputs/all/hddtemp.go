//go:build !custom || inputs || inputs.hddtemp

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/hddtemp" // register plugin
