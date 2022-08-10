//go:build !custom || inputs || inputs.opensmtpd

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/opensmtpd"
)
