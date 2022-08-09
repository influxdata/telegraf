//go:build all || inputs || inputs.opcua

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/opcua"
)
