//go:build !custom || inputs || inputs.ravendb

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/ravendb"
)
