//go:build all || inputs || inputs.openstack

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/openstack"
)
