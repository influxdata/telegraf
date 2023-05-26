//go:build !custom || serializers || serializers.wavefront

package all

import (
	_ "github.com/influxdata/telegraf/plugins/serializers/wavefront" // register plugin
)
