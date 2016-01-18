package all

import (
	_ "github.com/influxdb/telegraf/plugins/outputs/amon"
	_ "github.com/influxdb/telegraf/plugins/outputs/amqp"
        _ "github.com/influxdb/telegraf/plugins/outputs/cloudwatch"
	_ "github.com/influxdb/telegraf/plugins/outputs/datadog"
	_ "github.com/influxdb/telegraf/plugins/outputs/graphite"
	_ "github.com/influxdb/telegraf/plugins/outputs/influxdb"
	_ "github.com/influxdb/telegraf/plugins/outputs/kafka"
	_ "github.com/influxdb/telegraf/plugins/outputs/kinesis"
	_ "github.com/influxdb/telegraf/plugins/outputs/librato"
	_ "github.com/influxdb/telegraf/plugins/outputs/mqtt"
	_ "github.com/influxdb/telegraf/plugins/outputs/nsq"
	_ "github.com/influxdb/telegraf/plugins/outputs/opentsdb"
	_ "github.com/influxdb/telegraf/plugins/outputs/prometheus_client"
	_ "github.com/influxdb/telegraf/plugins/outputs/riemann"
)
