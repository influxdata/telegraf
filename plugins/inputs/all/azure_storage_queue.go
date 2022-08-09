//go:build all || inputs || inputs.azure_storage_queue

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/azure_storage_queue"
)
