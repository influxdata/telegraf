//go:build !custom || (migrations && (inputs || inputs.zookeeper))

package all

import _ "github.com/influxdata/telegraf/migrations/inputs_zookeeper" // register migration
