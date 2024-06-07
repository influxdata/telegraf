//go:build !custom || serializers || serializers.graphite

package all

import (
	_ "github.com/influxdata/telegraf/plugins/serializers/graphite" // register plugin
)
