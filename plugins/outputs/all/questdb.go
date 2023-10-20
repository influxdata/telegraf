//go:build !custom || outputs || outputs.questdb

package all

import _ "github.com/influxdata/telegraf/plugins/outputs/questdb" // register plugin
