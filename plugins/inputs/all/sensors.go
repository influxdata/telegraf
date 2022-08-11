//go:build !custom || inputs || inputs.sensors

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/sensors" // register plugin
