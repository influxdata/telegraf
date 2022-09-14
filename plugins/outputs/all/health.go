//go:build !custom || outputs || outputs.health

package all

import _ "github.com/influxdata/telegraf/plugins/outputs/health" // register plugin
