//go:build !custom || serializers || serializers.prometheusremotewrite

package all

import (
	_ "github.com/influxdata/telegraf/plugins/serializers/prometheusremotewrite" // register plugin
)
