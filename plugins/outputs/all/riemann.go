//go:build !custom || outputs || outputs.riemann

package all

import _ "github.com/influxdata/telegraf/plugins/outputs/riemann" // register plugin
