//go:build all || inputs || inputs.elasticsearch

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/elasticsearch"
)
