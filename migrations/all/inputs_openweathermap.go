//go:build !custom || (migrations && (inputs || inputs.openweathermap))

package all

import _ "github.com/influxdata/telegraf/migrations/inputs_openweathermap" // register migration
