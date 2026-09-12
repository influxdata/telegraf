//go:build !custom || inputs || inputs.openbsd_sensors

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/openbsd_sensors" // register plugin
