//go:build !custom || (migrations && (inputs || inputs.sqlserver))

package all

import _ "github.com/influxdata/telegraf/migrations/inputs_sqlserver" // register migration
