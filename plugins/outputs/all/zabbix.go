//go:build !custom || outputs || outputs.zabbix

package all

import _ "github.com/influxdata/telegraf/plugins/outputs/zabbix" // register plugin
