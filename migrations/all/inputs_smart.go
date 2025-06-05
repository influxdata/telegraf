//go:build !custom || (migrations && (inputs || inputs.smart))

package all

import _ "github.com/influxdata/telegraf/migrations/inputs_smart" // register migration
