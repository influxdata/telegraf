//go:build !custom || inputs || inputs.smartctl

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/smartctl" // register plugin
