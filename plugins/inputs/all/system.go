//go:build !custom || inputs || inputs.system

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/system" // register plugin
