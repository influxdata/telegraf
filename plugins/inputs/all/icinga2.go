//go:build !custom || inputs || inputs.icinga2

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/icinga2" // register plugin
