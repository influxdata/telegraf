//go:build !custom || inputs || inputs.openldap

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/openldap"
)
