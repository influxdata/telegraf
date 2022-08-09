//go:build all || inputs || inputs.nomad

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/nomad"
)
