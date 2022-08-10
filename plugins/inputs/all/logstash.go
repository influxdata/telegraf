//go:build !custom || inputs || inputs.logstash

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/logstash"
)
