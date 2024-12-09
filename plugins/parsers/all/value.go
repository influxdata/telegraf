//go:build !custom || parsers || parsers.value

package all

import _ "github.com/influxdata/telegraf/plugins/parsers/value" // register plugin
