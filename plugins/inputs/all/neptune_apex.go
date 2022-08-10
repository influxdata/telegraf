//go:build !custom || inputs || inputs.neptune_apex

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/neptune_apex"
)
