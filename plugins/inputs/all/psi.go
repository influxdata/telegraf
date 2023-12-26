//go:build !custom || inputs || inputs.psi

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/psi" // register plugin
