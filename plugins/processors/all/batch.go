//go:build !custom || processors || processors.batch

package all

import _ "github.com/influxdata/telegraf/plugins/processors/batch" // register plugin
