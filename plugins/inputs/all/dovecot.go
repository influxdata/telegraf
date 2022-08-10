//go:build !custom || inputs || inputs.dovecot

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/dovecot"
)
