//go:build !custom || (migrations && (inputs || inputs.elasticsearch))

package all

import _ "github.com/influxdata/telegraf/migrations/inputs_elasticsearch" // register migration
