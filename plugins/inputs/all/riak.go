//go:build !custom || inputs || inputs.riak

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/riak" // register plugin
