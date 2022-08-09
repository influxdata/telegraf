//go:build all || inputs || inputs.postfix

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/postfix"
)
