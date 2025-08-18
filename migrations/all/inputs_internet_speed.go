//go:build !custom || (migrations && (inputs || inputs.internet_speed))

package all

import _ "github.com/influxdata/telegraf/migrations/inputs_internet_speed" // register migration
