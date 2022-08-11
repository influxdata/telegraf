//go:build !custom || inputs || inputs.bond

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/bond" // register plugin
