//go:build !custom || outputs || outputs.exec

package all

import (
	_ "github.com/influxdata/telegraf/plugins/outputs/exec"
)
