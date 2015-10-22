package all

import (
	_ "github.com/influxdb/telegraf/outputs/amon"
	_ "github.com/influxdb/telegraf/outputs/amqp"
	_ "github.com/influxdb/telegraf/outputs/datadog"
	_ "github.com/influxdb/telegraf/outputs/influxdb"
	_ "github.com/influxdb/telegraf/outputs/kafka"
	_ "github.com/influxdb/telegraf/outputs/librato"
	_ "github.com/influxdb/telegraf/outputs/mqtt"
	_ "github.com/influxdb/telegraf/outputs/nsq"
	_ "github.com/influxdb/telegraf/outputs/opentsdb"
	_ "github.com/influxdb/telegraf/outputs/prometheus_client"
	_ "github.com/influxdb/telegraf/outputs/riemann"
)
