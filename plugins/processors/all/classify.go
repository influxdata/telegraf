//go:build !custom || processors || processors.classify

package all

import _ "github.com/influxdata/telegraf/plugins/processors/classify" // register plugin
