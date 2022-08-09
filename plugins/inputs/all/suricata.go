//go:build all || inputs || inputs.suricata

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/suricata"
)
