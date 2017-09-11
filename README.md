# Telegraf

Telegraf is an agent written in Go for collecting, processing, aggregating,
and writing metrics.

Telegraf is plugin-driven and has the concept of 4 distinct plugins:

1. [Input Plugins](#input-plugins) collect metrics from the system, services, or 3rd party APIs
2. [Processor Plugins](#processor-plugins) transform, decorate, and/or filter metrics
3. [Aggregator Plugins](#aggregator-plugins) create aggregate metrics (e.g. mean, min, max, quantiles, etc.)
4. [Output Plugins](#output-plugins) write metrics to various destinations

## Configuration and Installation
Please refer to the SignalFx [documentation](https://github.com/signalfx/integrations/tree/release/telegraf)
for more information on the installation and configuration of Telegraf.

## Advanced Configuration Options

See the [configuration guide](docs/CONFIGURATION.md) for a rundown of the more advanced
configuration options.

## Input Plugins

* [signalfx-metadata](./plugins/inputs/signalfx_metadata)
* [aerospike](./plugins/inputs/aerospike)
* [amqp_consumer](./plugins/inputs/amqp_consumer) (rabbitmq)
* [apache](./plugins/inputs/apache)
* [aws cloudwatch](./plugins/inputs/cloudwatch)
* [bcache](./plugins/inputs/bcache)
* [cassandra](./plugins/inputs/cassandra)
* [ceph](./plugins/inputs/ceph)
* [cgroup](./plugins/inputs/cgroup)
* [chrony](./plugins/inputs/chrony)
* [consul](./plugins/inputs/consul)
* [conntrack](./plugins/inputs/conntrack)
* [couchbase](./plugins/inputs/couchbase)
* [couchdb](./plugins/inputs/couchdb)
* [disque](./plugins/inputs/disque)
* [dmcache](./plugins/inputs/dmcache)
* [dns query time](./plugins/inputs/dns_query)
* [docker](./plugins/inputs/docker)
* [dovecot](./plugins/inputs/dovecot)
* [elasticsearch](./plugins/inputs/elasticsearch)
* [exec](./plugins/inputs/exec) (generic executable plugin, support JSON, influx, graphite and nagios)
* [fail2ban](./plugins/inputs/fail2ban)
* [filestat](./plugins/inputs/filestat)
* [fluentd](./plugins/inputs/fluentd)
* [graylog](./plugins/inputs/graylog)
* [haproxy](./plugins/inputs/haproxy)
* [hddtemp](./plugins/inputs/hddtemp)
* [http_response](./plugins/inputs/http_response)
* [httpjson](./plugins/inputs/httpjson) (generic JSON-emitting http service plugin)
* [internal](./plugins/inputs/internal)
* [influxdb](./plugins/inputs/influxdb)
* [interrupts](./plugins/inputs/interrupts)
* [ipmi_sensor](./plugins/inputs/ipmi_sensor)
* [iptables](./plugins/inputs/iptables)
* [jolokia](./plugins/inputs/jolokia)
* [kapacitor](./plugins/inputs/kapacitor)
* [kubernetes](./plugins/inputs/kubernetes)
* [leofs](./plugins/inputs/leofs)
* [lustre2](./plugins/inputs/lustre2)
* [mailchimp](./plugins/inputs/mailchimp)
* [memcached](./plugins/inputs/memcached)
* [mesos](./plugins/inputs/mesos)
* [minecraft](./plugins/inputs/minecraft)
* [mongodb](./plugins/inputs/mongodb)
* [mysql](./plugins/inputs/mysql)
* [net_response](./plugins/inputs/net_response)
* [nginx](./plugins/inputs/nginx)
* [nsq](./plugins/inputs/nsq)
* [nstat](./plugins/inputs/nstat)
* [ntpq](./plugins/inputs/ntpq)
* [openldap](./plugins/inputs/openldap)
* [phpfpm](./plugins/inputs/phpfpm)
* [phusion passenger](./plugins/inputs/passenger)
* [ping](./plugins/inputs/ping)
* [postgresql](./plugins/inputs/postgresql)
* [postgresql_extensible](./plugins/inputs/postgresql_extensible)
* [powerdns](./plugins/inputs/powerdns)
* [procstat](./plugins/inputs/procstat)
* [prometheus](./plugins/inputs/prometheus) (can be used for [Caddy server](./plugins/inputs/prometheus/README.md#usage-for-caddy-http-server))
* [puppetagent](./plugins/inputs/puppetagent)
* [rabbitmq](./plugins/inputs/rabbitmq)
* [raindrops](./plugins/inputs/raindrops)
* [redis](./plugins/inputs/redis)
* [rethinkdb](./plugins/inputs/rethinkdb)
* [riak](./plugins/inputs/riak)
* [salesforce](./plugins/inputs/salesforce)
* [sensors](./plugins/inputs/sensors)
* [snmp](./plugins/inputs/snmp)
* [snmp_legacy](./plugins/inputs/snmp_legacy)
* [sql server](./plugins/inputs/sqlserver) (microsoft)
* [tomcat](./plugins/inputs/tomcat)
* [twemproxy](./plugins/inputs/twemproxy)
* [varnish](./plugins/inputs/varnish)
* [zfs](./plugins/inputs/zfs)
* [zookeeper](./plugins/inputs/zookeeper)
* [win_perf_counters](./plugins/inputs/win_perf_counters) (windows performance counters)
* [win_services](./plugins/inputs/win_services)
* [sysstat](./plugins/inputs/sysstat)
* [system](./plugins/inputs/system)
    * cpu
    * mem
    * net
    * netstat
    * disk
    * diskio
    * swap
    * processes
    * kernel (/proc/stat)
    * kernel (/proc/vmstat)
    * linux_sysctl_fs (/proc/sys/fs)

Telegraf can also collect metrics via the following service plugins:

* [http_listener](./plugins/inputs/http_listener)
* [kafka_consumer](./plugins/inputs/kafka_consumer)
* [mqtt_consumer](./plugins/inputs/mqtt_consumer)
* [nats_consumer](./plugins/inputs/nats_consumer)
* [nsq_consumer](./plugins/inputs/nsq_consumer)
* [logparser](./plugins/inputs/logparser)
* [statsd](./plugins/inputs/statsd)
* [socket_listener](./plugins/inputs/socket_listener)
* [tail](./plugins/inputs/tail)
* [tcp_listener](./plugins/inputs/socket_listener)
* [udp_listener](./plugins/inputs/socket_listener)
* [webhooks](./plugins/inputs/webhooks)
  * [filestack](./plugins/inputs/webhooks/filestack)
  * [github](./plugins/inputs/webhooks/github)
  * [mandrill](./plugins/inputs/webhooks/mandrill)
  * [rollbar](./plugins/inputs/webhooks/rollbar)
  * [papertrail](./plugins/inputs/webhooks/papertrail)
* [zipkin](./plugins/inputs/zipkin)

Telegraf is able to parse the following input data formats into metrics, these
formats may be used with input plugins supporting the `data_format` option:

* [Collectd](./docs/DATA_FORMATS_INPUT.md#collectd)
* [Graphite](./docs/DATA_FORMATS_INPUT.md#graphite)
* [InfluxDB Line Protocol](./docs/DATA_FORMATS_INPUT.md#influx)
* [JSON](./docs/DATA_FORMATS_INPUT.md#json)
* [Nagios](./docs/DATA_FORMATS_INPUT.md#nagios)
* [Value](./docs/DATA_FORMATS_INPUT.md#value)

## Processor Plugins

* [printer](./plugins/processors/printer)

## Aggregator Plugins

* [signalfx_util](./plugins/aggregators/signalfx_util)
* [minmax](./plugins/aggregators/minmax)
* [histogram](./plugins/aggregators/histogram)

## Output Plugins

* [signalfx](./plugins/outputs/signalfx)
* [amon](./plugins/outputs/amon)
* [amqp](./plugins/outputs/amqp) (rabbitmq)
* [aws kinesis](./plugins/outputs/kinesis)
* [aws cloudwatch](./plugins/outputs/cloudwatch)
* [datadog](./plugins/outputs/datadog)
* [discard](./plugins/outputs/discard)
* [elasticsearch](./plugins/outputs/elasticsearch)
* [file](./plugins/outputs/file)
* [graphite](./plugins/outputs/graphite)
* [graylog](./plugins/outputs/graylog)
* [influxdb](./plugins/outputs/influxdb)
* [instrumental](./plugins/outputs/instrumental)
* [kafka](./plugins/outputs/kafka)
* [librato](./plugins/outputs/librato)
* [mqtt](./plugins/outputs/mqtt)
* [nats](./plugins/outputs/nats)
* [nsq](./plugins/outputs/nsq)
* [opentsdb](./plugins/outputs/opentsdb)
* [prometheus](./plugins/outputs/prometheus_client)
* [riemann](./plugins/outputs/riemann)
* [riemann_legacy](./plugins/outputs/riemann_legacy)
* [socket_writer](./plugins/outputs/socket_writer)
* [tcp](./plugins/outputs/socket_writer)
* [udp](./plugins/outputs/socket_writer)
