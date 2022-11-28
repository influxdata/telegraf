//go:build !custom || parsers || parsers.csv

package all

import _ "github.com/influxdata/telegraf/plugins/parsers/csv" // register plugin
