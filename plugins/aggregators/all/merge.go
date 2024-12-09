//go:build !custom || aggregators || aggregators.merge

package all

import _ "github.com/influxdata/telegraf/plugins/aggregators/merge" // register plugin
