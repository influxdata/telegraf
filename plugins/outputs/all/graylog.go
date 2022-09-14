//go:build !custom || outputs || outputs.graylog

package all

import _ "github.com/influxdata/telegraf/plugins/outputs/graylog" // register plugin
