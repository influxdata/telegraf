//go:build !custom || inputs || inputs.influxdb_listener

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/influxdb_listener" // register plugin
