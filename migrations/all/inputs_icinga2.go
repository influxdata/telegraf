//go:build !custom || (migrations && (inputs || inputs.icinga2))

package all

import _ "github.com/influxdata/telegraf/migrations/inputs_icinga2" // register migration
