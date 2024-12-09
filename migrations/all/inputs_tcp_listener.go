//go:build !custom || (migrations && (inputs || inputs.tcp_listener))

package all

import _ "github.com/influxdata/telegraf/migrations/inputs_tcp_listener" // register migration
