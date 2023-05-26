//go:build !custom || serializers || serializers.nowmetric

package all

import (
	_ "github.com/influxdata/telegraf/plugins/serializers/nowmetric" // register plugin
)
