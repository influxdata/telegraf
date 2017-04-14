# Telegraf [![Circle CI](https://circleci.com/gh/influxdata/telegraf.svg?style=svg)](https://circleci.com/gh/influxdata/telegraf) [![Docker pulls](https://img.shields.io/docker/pulls/library/telegraf.svg)](https://hub.docker.com/_/telegraf/)

Telegraf is an agent written in Go for collecting, processing, aggregating,
and writing metrics.

Design goals are to have a minimal memory footprint with a plugin system so
that developers in the community can easily add support for collecting metrics
from well known services (like Hadoop, Postgres, or Redis) and third party
APIs (like Mailchimp, AWS CloudWatch, or Google Analytics).

Telegraf is plugin-driven and has the concept of 4 distinct plugins:

1. [Input Plugins](#input-plugins) collect metrics from the system, services, or 3rd party APIs
2. [Processor Plugins](#processor-plugins) transform, decorate, and/or filter metrics
3. [Aggregator Plugins](#aggregator-plugins) create aggregate metrics (e.g. mean, min, max, quantiles, etc.)
4. [Output Plugins](#output-plugins) write metrics to various destinations

For more information on Processor and Aggregator plugins please [read this](./docs/AGGREGATORS_AND_PROCESSORS.md).

New plugins are designed to be easy to contribute,
we'll eagerly accept pull
requests and will manage the set of plugins that Telegraf supports.
See the [contributing guide](CONTRIBUTING.md) for instructions on writing
new plugins.

## Installation:

You can either download the binaries directly from the
[downloads](https://www.influxdata.com/downloads) page.

A few alternate installs are available here as well:

### FreeBSD tarball:

Latest:
* https://dl.influxdata.com/telegraf/releases/telegraf-VERSION_freebsd_amd64.tar.gz

### Ansible Role:

Ansible role: https://github.com/rossmcdonald/telegraf

### From Source:

Telegraf manages dependencies via [gdm](https://github.com/sparrc/gdm),
which gets installed via the Makefile
if you don't have it already. You also must build with golang version 1.8+.

1. [Install Go](https://golang.org/doc/install)
2. [Setup your GOPATH](https://golang.org/doc/code.html#GOPATH)
3. Run `go get github.com/influxdata/telegraf`
4. Run `cd $GOPATH/src/github.com/influxdata/telegraf`
5. Run `make`

## How to use it:

See usage with:

```
telegraf --help
```

#### Generate a telegraf config file:

```
telegraf config > telegraf.conf
```

#### Generate config with only cpu input & influxdb output plugins defined

```
telegraf --input-filter cpu --output-filter influxdb config
```

#### Run a single telegraf collection, outputing metrics to stdout

```
telegraf --config telegraf.conf -test
```

#### Run telegraf with all plugins defined in config file

```
telegraf --config telegraf.conf
```

#### Run telegraf, enabling the cpu & memory input, and influxdb output plugins

```
telegraf --config telegraf.conf -input-filter cpu:mem -output-filter influxdb
```


## Configuration

See the [configuration guide](docs/CONFIGURATION.md) for a rundown of the more advanced
configuration options.

## Supported Input Plugins

Telegraf currently has support for collecting metrics from many sources. For
more information on each, please look at the directory of the same name in
`plugins/inputs`.

Currently implemented sources:

* [aws cloudwatch](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/cloudwatch)
* [aerospike](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/aerospike)
* [apache](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/apache)
* [bcache](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/bcache)
* [beanstalkd](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/beanstalkd)
* [cassandra](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/cassandra)
* [ceph](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/ceph)
* [chrony](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/chrony)
* [consul](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/consul)
* [conntrack](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/conntrack)
* [couchbase](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/couchbase)
* [couchdb](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/couchdb)
* [disque](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/disque)
* [dns query time](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/dns_query)
* [docker](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/docker)
* [dovecot](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/dovecot)
* [elasticsearch](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/elasticsearch)
* [exec](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/exec) (generic executable plugin, support JSON, influx, graphite and nagios)
* [filestat](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/filestat)
* [haproxy](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/haproxy)
* [hddtemp](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/hddtemp)
* [http_response](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/http_response)
* [httpjson](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/httpjson) (generic JSON-emitting http service plugin)
* [influxdb](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/influxdb)
* [ipmi_sensor](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/ipmi_sensor)
* [iptables](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/iptables)
* [jolokia](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/jolokia)
* [leofs](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/leofs)
* [lustre2](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/lustre2)
* [mailchimp](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/mailchimp)
* [memcached](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/memcached)
* [mesos](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/mesos)
* [mongodb](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/mongodb)
* [mysql](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/mysql)
* [net_response](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/net_response)
* [nginx](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/nginx)
* [nsq](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/nsq)
* [nstat](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/nstat)
* [ntpq](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/ntpq)
* [phpfpm](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/phpfpm)
* [phusion passenger](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/passenger)
* [ping](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/ping)
* [postgresql](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/postgresql)
* [postgresql_extensible](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/postgresql_extensible)
* [powerdns](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/powerdns)
* [procstat](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/procstat)
* [prometheus](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/prometheus)
* [puppetagent](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/puppetagent)
* [rabbitmq](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/rabbitmq)
* [raindrops](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/raindrops)
* [redis](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/redis)
* [rethinkdb](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/rethinkdb)
* [riak](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/riak)
* [sensors](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/sensors)
* [snmp](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/snmp)
* [snmp_legacy](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/snmp_legacy)
* [sql server](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/sqlserver) (microsoft)
* [twemproxy](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/twemproxy)
* [varnish](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/varnish)
* [zfs](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/zfs)
* [zookeeper](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/zookeeper)
* [win_perf_counters ](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/win_perf_counters) (windows performance counters)
* [sysstat](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/sysstat)
* [system](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/system)
## Input Plugins

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
* [dns query time](./plugins/inputs/dns_query)
* [docker](./plugins/inputs/docker)
* [dovecot](./plugins/inputs/dovecot)
* [elasticsearch](./plugins/inputs/elasticsearch)
* [exec](./plugins/inputs/exec) (generic executable plugin, support JSON, influx, graphite and nagios)
* [filestat](./plugins/inputs/filestat)
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
* [kubernetes](./plugins/inputs/kubernetes)
* [leofs](./plugins/inputs/leofs)
* [lustre2](./plugins/inputs/lustre2)
* [mailchimp](./plugins/inputs/mailchimp)
* [memcached](./plugins/inputs/memcached)
* [mesos](./plugins/inputs/mesos)
* [mongodb](./plugins/inputs/mongodb)
* [mysql](./plugins/inputs/mysql)
* [net_response](./plugins/inputs/net_response)
* [nginx](./plugins/inputs/nginx)
* [nsq](./plugins/inputs/nsq)
* [nstat](./plugins/inputs/nstat)
* [ntpq](./plugins/inputs/ntpq)
* [phpfpm](./plugins/inputs/phpfpm)
* [phusion passenger](./plugins/inputs/passenger)
* [ping](./plugins/inputs/ping)
* [postgresql](./plugins/inputs/postgresql)
* [postgresql_extensible](./plugins/inputs/postgresql_extensible)
* [powerdns](./plugins/inputs/powerdns)
* [procstat](./plugins/inputs/procstat)
* [prometheus](./plugins/inputs/prometheus)
* [puppetagent](./plugins/inputs/puppetagent)
* [rabbitmq](./plugins/inputs/rabbitmq)
* [raindrops](./plugins/inputs/raindrops)
* [redis](./plugins/inputs/redis)
* [rethinkdb](./plugins/inputs/rethinkdb)
* [riak](./plugins/inputs/riak)
* [sensors](./plugins/inputs/sensors)
* [snmp](./plugins/inputs/snmp)
* [snmp_legacy](./plugins/inputs/snmp_legacy)
* [sql server](./plugins/inputs/sqlserver) (microsoft)
* [twemproxy](./plugins/inputs/twemproxy)
* [varnish](./plugins/inputs/varnish)
* [zfs](./plugins/inputs/zfs)
* [zookeeper](./plugins/inputs/zookeeper)
* [win_perf_counters ](./plugins/inputs/win_perf_counters) (windows performance counters)
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

Telegraf is able to parse the following input data formats into metrics, these
formats may be used with input plugins supporting the `data_format` option:

* [InfluxDB Line Protocol](./docs/DATA_FORMATS_INPUT.md#influx)
* [JSON](./docs/DATA_FORMATS_INPUT.md#json)
* [Graphite](./docs/DATA_FORMATS_INPUT.md#graphite)
* [Value](./docs/DATA_FORMATS_INPUT.md#value)
* [Nagios](./docs/DATA_FORMATS_INPUT.md#nagios)
* [Collectd](./docs/DATA_FORMATS_INPUT.md#collectd)

## Processor Plugins

* [printer](./plugins/processors/printer)

## Aggregator Plugins

* [minmax](./plugins/aggregators/minmax)

## Output Plugins

* [influxdb](./plugins/outputs/influxdb)
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

## Contributing

Please see the
[contributing guide](CONTRIBUTING.md)
for details on contributing a plugin to Telegraf.
