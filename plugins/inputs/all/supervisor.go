//go:build !custom || inputs || inputs.supervisor

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/supervisor" // register plugin
