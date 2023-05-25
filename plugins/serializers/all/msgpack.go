//go:build !custom || serializers || serializers.msgpack

package all

import (
	_ "github.com/influxdata/telegraf/plugins/serializers/msgpack" // register plugin
)
