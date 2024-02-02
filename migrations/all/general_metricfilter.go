//go:build !custom || migrations

package all

import _ "github.com/influxdata/telegraf/migrations/general_metricfilter" // register migration
