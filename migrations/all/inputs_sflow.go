//go:build !custom || (migrations && (inputs || inputs.sflow))

package all

import _ "github.com/influxdata/telegraf/migrations/inputs_sflow" // register migration
