//go:build !custom || inputs || inputs.nsq

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/nsq"
)
