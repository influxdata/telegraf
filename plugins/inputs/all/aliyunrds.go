//go:build !custom || inputs || inputs.aliyunrds

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/aliyunrds" // register plugin
