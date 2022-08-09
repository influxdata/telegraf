//go:build all || inputs || inputs.nfsclient

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/nfsclient"
)
