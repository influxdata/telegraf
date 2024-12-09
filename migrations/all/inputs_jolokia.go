//go:build !custom || (migrations && (inputs || inputs.jolokia))

package all

import _ "github.com/influxdata/telegraf/migrations/inputs_jolokia" // register migration
