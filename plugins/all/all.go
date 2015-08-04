package all

import (
	_ "github.com/influxdb/telegraf/plugins/disque"
	_ "github.com/influxdb/telegraf/plugins/elasticsearch"
	_ "github.com/influxdb/telegraf/plugins/kafka_consumer"
	_ "github.com/influxdb/telegraf/plugins/lustre2"
	_ "github.com/influxdb/telegraf/plugins/memcached"
	_ "github.com/influxdb/telegraf/plugins/mongodb"
	_ "github.com/influxdb/telegraf/plugins/mysql"
	_ "github.com/influxdb/telegraf/plugins/postgresql"
	_ "github.com/influxdb/telegraf/plugins/prometheus"
	_ "github.com/influxdb/telegraf/plugins/rabbitmq"
	_ "github.com/influxdb/telegraf/plugins/redis"
	_ "github.com/influxdb/telegraf/plugins/rethinkdb"
	_ "github.com/influxdb/telegraf/plugins/system"
)
