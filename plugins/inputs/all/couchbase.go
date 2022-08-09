//go:build all || inputs || inputs.couchbase

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/couchbase"
)
