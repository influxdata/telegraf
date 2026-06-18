//go:build !custom || (migrations && (inputs || inputs.aerospike))

package all

import _ "github.com/influxdata/telegraf/migrations/inputs_aerospike" // register migration
