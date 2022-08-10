//go:build !custom || outputs || outputs.socket_writer

package all

import (
	_ "github.com/influxdata/telegraf/plugins/outputs/socket_writer"
)
