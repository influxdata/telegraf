//go:build !custom || inputs || inputs.leofs

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/leofs" // register plugin
