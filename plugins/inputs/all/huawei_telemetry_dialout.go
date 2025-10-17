//go:build !custom || inputs || inputs.huawei_telemetry_dialout

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/huawei_telemetry_dialout" // register plugin
