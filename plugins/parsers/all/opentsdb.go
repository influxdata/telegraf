//go:build !custom || parsers || parsers.opentsdb

package all

import _ "github.com/influxdata/telegraf/plugins/parsers/opentsdb" // register plugin
