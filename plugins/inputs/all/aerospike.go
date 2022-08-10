//go:build !custom || inputs || inputs.aerospike

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/aerospike"
)
