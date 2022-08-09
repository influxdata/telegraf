//go:build all || inputs || inputs.dcos

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/dcos"
)
