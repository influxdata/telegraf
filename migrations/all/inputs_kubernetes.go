//go:build !custom || (migrations && (inputs || inputs.kubernetes))

package all

import _ "github.com/influxdata/telegraf/migrations/inputs_kubernetes" // register migration
