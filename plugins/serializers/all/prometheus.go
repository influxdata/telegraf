//go:build !custom || serializers || serializers.prometheus

package all

import (
	_ "github.com/influxdata/telegraf/plugins/serializers/prometheus" // register plugin
)
