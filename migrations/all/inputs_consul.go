//go:build !custom || (migrations && (inputs || inputs.consul))

package all

import _ "github.com/influxdata/telegraf/migrations/inputs_consul" // register migration
