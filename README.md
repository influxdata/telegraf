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

### Linux deb and rpm Packages:

Latest:
* https://dl.influxdata.com/telegraf/releases/telegraf_1.0.1_amd64.deb
* https://dl.influxdata.com/telegraf/releases/telegraf-1.0.1.x86_64.rpm

Latest (arm):
* https://dl.influxdata.com/telegraf/releases/telegraf_1.0.1_armhf.deb
* https://dl.influxdata.com/telegraf/releases/telegraf-1.0.1.armhf.rpm

##### Package Instructions:

* Telegraf binary is installed in `/usr/bin/telegraf`
* Telegraf daemon configuration file is in `/etc/telegraf/telegraf.conf`
* On sysv systems, the telegraf daemon can be controlled via
`service telegraf [action]`
* On systemd systems (such as Ubuntu 15+), the telegraf daemon can be
controlled via `systemctl [action] telegraf`

### yum/apt Repositories:

There is a yum/apt repo available for the whole InfluxData stack, see
[here](https://docs.influxdata.com/influxdb/latest/introduction/installation/#installation)
for instructions on setting up the repo. Once it is configured, you will be able
to use this repo to install & update telegraf.

### Linux tarballs:

Latest:
* https://dl.influxdata.com/telegraf/releases/telegraf-1.0.1_linux_amd64.tar.gz
* https://dl.influxdata.com/telegraf/releases/telegraf-1.0.1_linux_i386.tar.gz
* https://dl.influxdata.com/telegraf/releases/telegraf-1.0.1_linux_armhf.tar.gz

### FreeBSD tarball:

Latest:
* https://dl.influxdata.com/telegraf/releases/telegraf-1.0.1_freebsd_amd64.tar.gz

### Ansible Role:

Ansible role: https://github.com/rossmcdonald/telegraf

### OSX via Homebrew:

```
brew update
brew install telegraf
```

### Windows Binaries (EXPERIMENTAL)

Latest:
* https://dl.influxdata.com/telegraf/releases/telegraf-1.0.1_windows_amd64.zip

### From Source:

Telegraf manages dependencies via [gdm](https://github.com/sparrc/gdm),
which gets installed via the Makefile
if you don't have it already. You also must build with golang version 1.5+.

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

### Generate a telegraf config file:

```
telegraf config > telegraf.conf
```

### Generate config with only cpu input & influxdb output plugins defined

```
telegraf --input-filter cpu --output-filter influxdb config
```

### Run a single telegraf collection, outputing metrics to stdout

```
telegraf --config telegraf.conf -test
```

### Run telegraf with all plugins defined in config file

```
telegraf --config telegraf.conf
```

### Run telegraf, enabling the cpu & memory input, and influxdb output plugins

```
telegraf --config telegraf.conf -input-filter cpu:mem -output-filter influxdb
```


## Configuration

See the [configuration guide](docs/CONFIGURATION.md) for a rundown of the more advanced
configuration options.

## Input Plugins

* [aws cloudwatch](./plugins/inputs/cloudwatch)
* [aerospike](./plugins/inputs/aerospike)
* [apache](./plugins/inputs/apache)
* [bcache](./plugins/inputs/bcache)
* [cassandra](./plugins/inputs/cassandra)
* [ceph](./plugins/inputs/ceph)
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
* [ipmi_sensor](./plugins/inputs/ipmi_sensor)
* [iptables](./plugins/inputs/iptables)
* [jolokia](./plugins/inputs/jolokia)
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

Telegraf can also collect metrics via the following service plugins:

* [http_listener](./plugins/inputs/http_listener)
* [kafka_consumer](./plugins/inputs/kafka_consumer)
* [mqtt_consumer](./plugins/inputs/mqtt_consumer)
* [nats_consumer](./plugins/inputs/nats_consumer)
* [nsq_consumer](./plugins/inputs/nsq_consumer)
* [logparser](./plugins/inputs/logparser)
* [statsd](./plugins/inputs/statsd)
* [tail](./plugins/inputs/tail)
* [tcp_listener](./plugins/inputs/tcp_listener)
* [udp_listener](./plugins/inputs/udp_listener)
* [webhooks](./plugins/inputs/webhooks)
  * [filestack](./plugins/inputs/webhooks/filestack)
  * [github](./plugins/inputs/webhooks/github)
  * [mandrill](./plugins/inputs/webhooks/mandrill)
  * [rollbar](./plugins/inputs/webhooks/rollbar)

## Processor Plugins

* [printer](./plugins/processors/printer)

## Aggregator Plugins

* [minmax](./plugins/aggregators/minmax)

## Output Plugins

* [influxdb](./plugins/outputs/influxdb)
* [amon](./plugins/outputs/amon)
* [amqp](./plugins/outputs/amqp)
* [aws kinesis](./plugins/outputs/kinesis)
* [aws cloudwatch](./plugins/outputs/cloudwatch)
* [datadog](./plugins/outputs/datadog)
* [discard](./plugins/outputs/discard)
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

## Contributing

Please see the
[contributing guide](CONTRIBUTING.md)
for details on contributing a plugin to Telegraf.
