//go:build !custom || (migrations && (inputs || inputs.ntpq))

package all

import _ "github.com/influxdata/telegraf/migrations/inputs_ntpq" // register migration
