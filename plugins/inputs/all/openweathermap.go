//go:build !custom || inputs || inputs.openweathermap

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/openweathermap"
)
