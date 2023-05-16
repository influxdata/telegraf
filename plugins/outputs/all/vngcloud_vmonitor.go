//go:build !custom || outputs || outputs.vngcloud_vmonitor

package all

import _ "github.com/influxdata/telegraf/plugins/outputs/vngcloud_vmonitor" // register plugin