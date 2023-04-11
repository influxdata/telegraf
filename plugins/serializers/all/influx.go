//go:build !custom || serializers || serializers.influx

package all

import (
	_ "github.com/influxdata/telegraf/plugins/serializers/influx" // register plugin
)
