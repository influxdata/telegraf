//go:build !custom || (migrations && (inputs || inputs.http_listener))

package all

import _ "github.com/influxdata/telegraf/migrations/inputs_http_listener" // register migration
