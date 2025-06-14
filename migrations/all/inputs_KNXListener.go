//go:build !custom || (migrations && (inputs || inputs.KNXListener))

package all

import _ "github.com/influxdata/telegraf/migrations/inputs_KNXListener" // register migration
