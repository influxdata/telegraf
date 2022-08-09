//go:build all || inputs || inputs.zfs

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/zfs"
)
