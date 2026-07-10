//go:build !custom || parsers || parsers.concat

package all

import _ "github.com/influxdata/telegraf/plugins/parsers/concat" // register plugin
