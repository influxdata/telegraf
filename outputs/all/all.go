package all

import (
	_ "github.com/influxdb/telegraf/outputs/amqp"
	_ "github.com/influxdb/telegraf/outputs/datadog"
	_ "github.com/influxdb/telegraf/outputs/influxdb"
	_ "github.com/influxdb/telegraf/outputs/kafka"
	_ "github.com/influxdb/telegraf/outputs/mqtt"
	_ "github.com/influxdb/telegraf/outputs/opentsdb"
)
