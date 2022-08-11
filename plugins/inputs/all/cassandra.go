//go:build !custom || inputs || inputs.cassandra

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/cassandra" // register plugin
