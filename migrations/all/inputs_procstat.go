//go:build !custom || (migrations && (inputs || inputs.procstat))

package all

import _ "github.com/influxdata/telegraf/migrations/inputs_procstat" // register migration
