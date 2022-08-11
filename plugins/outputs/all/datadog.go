//go:build !custom || outputs || outputs.datadog

package all

import _ "github.com/influxdata/telegraf/plugins/outputs/datadog" // register plugin
