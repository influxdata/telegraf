//go:build all || inputs || inputs.couchdb

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/couchdb"
)
