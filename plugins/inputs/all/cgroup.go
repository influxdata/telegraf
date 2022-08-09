//go:build all || inputs || inputs.cgroup

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/cgroup"
)
