//go:build !custom || (migrations && (outputs || outputs.librato))

package all

import _ "github.com/influxdata/telegraf/migrations/outputs_librato" // register migration
