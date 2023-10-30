//go:build !custom || (migrations && (outputs || outputs.riemann_legacy))

package all

import _ "github.com/influxdata/telegraf/migrations/outputs_riemann_legacy" // register migration
