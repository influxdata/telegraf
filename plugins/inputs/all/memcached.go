//go:build !custom || inputs || inputs.memcached

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/memcached"
)
