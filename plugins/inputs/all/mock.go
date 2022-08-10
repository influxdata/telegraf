//go:build !custom || inputs || inputs.mock

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/mock"
)
