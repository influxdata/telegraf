//go:build !custom || outputs || outputs.wavefront

package all

import _ "github.com/influxdata/telegraf/plugins/outputs/wavefront" // register plugin
