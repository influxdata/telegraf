//go:build !custom || outputs || outputs.azure_data_explorer

package all

import (
	_ "github.com/influxdata/telegraf/plugins/outputs/azure_data_explorer"
)
