//go:build !custom || processors || processors.scaler

package all

import _ "github.com/influxdata/telegraf/plugins/processors/scaler" // register plugin
