//go:build !custom || parsers || parsers.binary

package all

import _ "github.com/influxdata/telegraf/plugins/parsers/binary" // register plugin
