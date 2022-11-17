//go:build !custom || parsers || parsers.dropwizard

package all

import _ "github.com/influxdata/telegraf/plugins/parsers/dropwizard" // register plugin
