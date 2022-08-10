//go:build !custom || inputs || inputs.clickhouse

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/clickhouse"
)
