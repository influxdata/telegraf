//go:build !custom || inputs || inputs.vsphere

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/vsphere"
)
