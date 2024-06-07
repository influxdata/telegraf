//go:build !custom || serializers || serializers.csv

package all

import (
	_ "github.com/influxdata/telegraf/plugins/serializers/csv" // register plugin
)
