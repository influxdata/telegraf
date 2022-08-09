//go:build all || inputs || inputs.kubernetes

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/kubernetes"
)
