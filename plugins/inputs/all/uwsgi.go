//go:build all || inputs || inputs.uwsgi

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/uwsgi"
)
