//go:build !custom || inputs || inputs.influxdb_v3_listener

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/influxdb_v3_listener" // register plugin
