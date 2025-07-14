//go:build !custom || (migrations && (outputs || outputs.remotefile))

package all

import _ "github.com/influxdata/telegraf/migrations/outputs_remotefile" // register migration
