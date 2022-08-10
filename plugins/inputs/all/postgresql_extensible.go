//go:build !custom || inputs || inputs.postgresql_extensible

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/postgresql_extensible"
)
