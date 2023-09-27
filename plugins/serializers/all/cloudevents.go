//go:build !custom || serializers || serializers.cloudevents

package all

import (
	_ "github.com/influxdata/telegraf/plugins/serializers/cloudevents" // register plugin
)
