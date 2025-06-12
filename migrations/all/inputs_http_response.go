//go:build !custom || (migrations && (inputs || inputs.http_response))

package all

import _ "github.com/influxdata/telegraf/migrations/inputs_http_response" // register migration
