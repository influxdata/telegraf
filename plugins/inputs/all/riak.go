//go:build all || inputs || inputs.riak

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/riak"
)
