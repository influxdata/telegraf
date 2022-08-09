//go:build all || inputs || inputs.mysql

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/mysql"
)
