//go:build all || inputs || inputs.http_response

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/http_response"
)
