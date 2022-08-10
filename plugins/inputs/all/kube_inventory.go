//go:build !custom || inputs || inputs.kube_inventory

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/kube_inventory"
)
