//go:build !custom || inputs || inputs.journald

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/journald" // register plugin
