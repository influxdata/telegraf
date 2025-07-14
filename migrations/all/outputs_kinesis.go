//go:build !custom || (migrations && (outputs || outputs.kinesis))

package all

import _ "github.com/influxdata/telegraf/migrations/outputs_kinesis" // register migration
