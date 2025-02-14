//go:build !custom || parsers || parsers.openmetrics

package all

import _ "github.com/influxdata/telegraf/plugins/parsers/openmetrics" // register plugin
