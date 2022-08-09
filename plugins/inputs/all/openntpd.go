//go:build all || inputs || inputs.openntpd

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/openntpd"
)
