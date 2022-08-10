//go:build !custom || inputs || inputs.dns_query

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/dns_query"
)
