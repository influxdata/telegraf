//go:build !custom || serializers || serializers.splunkmetric

package all

import (
	_ "github.com/influxdata/telegraf/plugins/serializers/splunkmetric" // register plugin
)
