//go:build !custom || processors || processors.topk

package all

import _ "github.com/influxdata/telegraf/plugins/processors/topk" // register plugin
