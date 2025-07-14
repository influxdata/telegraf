//go:build !custom || (migrations && (inputs || inputs.http_listener_v2))

package all

import _ "github.com/influxdata/telegraf/migrations/inputs_http_listener_v2" // register migration
