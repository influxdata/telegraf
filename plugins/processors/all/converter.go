//go:build !custom || processors || processors.converter

package all

import _ "github.com/influxdata/telegraf/plugins/processors/converter" // register plugin
