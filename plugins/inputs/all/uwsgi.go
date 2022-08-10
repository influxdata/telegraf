//go:build !custom || inputs || inputs.uwsgi

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/uwsgi"
)
