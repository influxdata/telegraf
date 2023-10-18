//go:build !custom || (migrations && (inputs || inputs.disk))

package all

import _ "github.com/influxdata/telegraf/migrations/inputs_disk" // register migration
