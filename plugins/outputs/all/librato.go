//go:build !custom || outputs || outputs.librato

package all

import _ "github.com/influxdata/telegraf/plugins/outputs/librato" // register plugin
