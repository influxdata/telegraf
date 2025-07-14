//go:build !custom || (migrations && (inputs || inputs.nsq_consumer))

package all

import _ "github.com/influxdata/telegraf/migrations/inputs_nsq_consumer" // register migration
