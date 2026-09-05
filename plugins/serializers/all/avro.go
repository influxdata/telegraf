//go:build !custom || serializers || serializers.avro

package all

import (
	_ "github.com/influxdata/telegraf/plugins/serializers/avro" // register plugin
)
