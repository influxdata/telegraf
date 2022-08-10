//go:build !custom || inputs || inputs.kubernetes

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/kubernetes"
)
