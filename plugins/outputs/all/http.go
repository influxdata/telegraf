//go:build !custom || outputs || outputs.http

package all

import (
	_ "github.com/influxdata/telegraf/plugins/outputs/http"
)
