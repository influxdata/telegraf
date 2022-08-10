//go:build !custom || inputs || inputs.redfish

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/redfish"
)
