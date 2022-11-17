//go:build !custom || parsers || parsers.collectd

package all

import _ "github.com/influxdata/telegraf/plugins/parsers/collectd" // register plugin
