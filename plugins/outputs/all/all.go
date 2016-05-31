package all

import (
	_ "github.com/influxdata/telegraf/plugins/outputs/amon"
	_ "github.com/influxdata/telegraf/plugins/outputs/amqp"
	_ "github.com/influxdata/telegraf/plugins/outputs/cloudwatch"
	_ "github.com/influxdata/telegraf/plugins/outputs/datadog"
	_ "github.com/influxdata/telegraf/plugins/outputs/file"
	_ "github.com/influxdata/telegraf/plugins/outputs/graphite"
	_ "github.com/influxdata/telegraf/plugins/outputs/influxdb"
	_ "github.com/influxdata/telegraf/plugins/outputs/instrumental"
	_ "github.com/influxdata/telegraf/plugins/outputs/kafka"
	_ "github.com/influxdata/telegraf/plugins/outputs/kinesis"
	_ "github.com/influxdata/telegraf/plugins/outputs/librato"
	_ "github.com/influxdata/telegraf/plugins/outputs/mqtt"
	_ "github.com/influxdata/telegraf/plugins/outputs/nsq"
	_ "github.com/influxdata/telegraf/plugins/outputs/opentsdb"
	_ "github.com/influxdata/telegraf/plugins/outputs/prometheus_client"
	_ "github.com/influxdata/telegraf/plugins/outputs/riemann"
)
