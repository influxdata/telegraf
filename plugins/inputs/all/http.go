//go:build all || inputs || inputs.http

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/http"
)
