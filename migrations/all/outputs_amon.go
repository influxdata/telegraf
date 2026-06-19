//go:build !custom || (migrations && (outputs || outputs.amon))

package all

import _ "github.com/influxdata/telegraf/migrations/outputs_amon" // register migration
