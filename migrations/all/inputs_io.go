//go:build !custom || (migrations && (inputs || inputs.io || inputs.diskio))

package all

import _ "github.com/influxdata/telegraf/migrations/inputs_io" // register migration
