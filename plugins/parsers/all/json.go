//go:build !custom || parsers || parsers.json

package all

import _ "github.com/influxdata/telegraf/plugins/parsers/json" // register plugin
