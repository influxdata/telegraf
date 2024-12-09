//go:build !custom || serializers || serializers.carbon2

package all

import (
	_ "github.com/influxdata/telegraf/plugins/serializers/carbon2" // register plugin
)
