//go:build !custom || inputs || inputs.nfsclient

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/nfsclient"
)
