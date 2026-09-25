//go:build !custom || inputs || inputs.fxmacrodata

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/fxmacrodata" // register plugin
