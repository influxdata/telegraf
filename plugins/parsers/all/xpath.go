//go:build !custom || parsers || parsers.xpath

package all

import _ "github.com/influxdata/telegraf/plugins/parsers/xpath" // register plugin
