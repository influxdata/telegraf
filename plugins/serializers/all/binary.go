//go:build !custom || serializers || serializers.binary

package all

import (
	_ "github.com/influxdata/telegraf/plugins/serializers/binary" // register plugin
)
