package filtered

import (
	_ "github.com/influxdata/telegraf/plugins/outputs/discard"
	_ "github.com/influxdata/telegraf/plugins/outputs/file"
	_ "github.com/influxdata/telegraf/plugins/outputs/http"
	_ "github.com/influxdata/telegraf/plugins/outputs/influxdb"
	_ "github.com/influxdata/telegraf/plugins/outputs/influxdb_v2"
	_ "github.com/influxdata/telegraf/plugins/outputs/kafka"
	_ "github.com/influxdata/telegraf/plugins/outputs/mqtt"
	_ "github.com/influxdata/telegraf/plugins/outputs/nsq"
	_ "github.com/influxdata/telegraf/plugins/outputs/opentsdb"
	_ "github.com/influxdata/telegraf/plugins/outputs/socket_writer"
	_ "github.com/influxdata/telegraf/plugins/outputs/stackdriver"
)
