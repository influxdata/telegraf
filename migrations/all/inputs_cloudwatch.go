//go:build !custom || (migrations && (inputs || inputs.cloudwatch))

package all

import _ "github.com/influxdata/telegraf/migrations/inputs_cloudwatch" // register migration
