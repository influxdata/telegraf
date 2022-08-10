//go:build !custom || inputs || inputs.nvidia_smi

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/nvidia_smi"
)
