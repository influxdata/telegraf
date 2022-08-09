//go:build all || inputs || inputs.clickhouse

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/clickhouse"
)
