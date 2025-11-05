//go:build !custom || inputs || inputs.timex

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/timex" // register plugin
