//go:build !custom || outputs || outputs.postgresql

package all

import _ "github.com/influxdata/telegraf/plugins/outputs/postgresql" // register plugin
