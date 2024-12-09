//go:build !custom || parsers || parsers.prometheusremotewrite

package all

import _ "github.com/influxdata/telegraf/plugins/parsers/prometheusremotewrite" // register plugin
