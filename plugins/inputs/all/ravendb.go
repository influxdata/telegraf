//go:build all || inputs || inputs.ravendb

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/ravendb"
)
