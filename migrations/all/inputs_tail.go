//go:build !custom || (migrations && (inputs || inputs.tail))

package all

import _ "github.com/influxdata/telegraf/migrations/inputs_tail" // register migration
