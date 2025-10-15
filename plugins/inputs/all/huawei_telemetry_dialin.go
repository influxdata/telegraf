//go:build !custom || inputs || inputs.huawei_telemetry_dialin

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/huawei_telemetry_dialin" // register plugin
