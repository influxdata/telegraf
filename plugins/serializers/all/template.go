//go:build !custom || serializers || serializers.template

package all

import (
	_ "github.com/influxdata/telegraf/plugins/serializers/template" // register plugin
)
