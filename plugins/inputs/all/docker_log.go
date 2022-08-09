//go:build all || inputs || inputs.docker_log

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/docker_log"
)
