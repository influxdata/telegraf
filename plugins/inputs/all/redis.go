//go:build all || inputs || inputs.redis

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/redis"
)
