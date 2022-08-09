//go:build all || inputs || inputs.postgresql

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/postgresql"
)
