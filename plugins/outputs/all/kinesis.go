//go:build !custom || outputs || outputs.kinesis

package all

import _ "github.com/influxdata/telegraf/plugins/outputs/kinesis" // register plugin
