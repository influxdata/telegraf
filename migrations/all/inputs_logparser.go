//go:build !custom || (migrations && (inputs || inputs.logparser))

package all

import _ "github.com/influxdata/telegraf/migrations/inputs_logparser" // register migration
