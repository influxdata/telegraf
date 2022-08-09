//go:build all || inputs || inputs.ceph

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/ceph"
)
