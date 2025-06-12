//go:build !custom || (migrations && (inputs || inputs.http))

package all

import _ "github.com/influxdata/telegraf/migrations/inputs_http" // register migration
