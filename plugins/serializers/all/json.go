//go:build !custom || serializers || serializers.json

package all

import (
	_ "github.com/influxdata/telegraf/plugins/serializers/json" // register plugin
)
