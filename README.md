# Telegraf [![Circle CI](https://circleci.com/gh/influxdata/telegraf.svg?style=svg)](https://circleci.com/gh/influxdata/telegraf) [![Docker pulls](https://img.shields.io/docker/pulls/library/telegraf.svg)](https://hub.docker.com/_/telegraf/)

Telegraf is an agent written in Go for collecting, processing, aggregating,
and writing metrics.

Design goals are to have a minimal memory footprint with a plugin system so
that developers in the community can easily add support for collecting metrics
.  For an example configuration referencet from local or remote services.

Telegraf is plugin-driven and has the concept of 4 distinct plugins:

1. [Input Plugins](#input-plugins) collect metrics from the system, services, or 3rd party APIs
2. [Processor Plugins](#processor-plugins) transform, decorate, and/or filter metrics
3. [Aggregator Plugins](#aggregator-plugins) create aggregate metrics (e.g. mean, min, max, quantiles, etc.)
4. [Output Plugins](#output-plugins) write metrics to various destinations

For more information on Processor and Aggregator plugins please [read this](./docs/AGGREGATORS_AND_PROCESSORS.md).

New plugins are designed to be easy to contribute,
we'll eagerly accept pull
requests and will manage the set of plugins that Telegraf supports.

## Contributing

There are many ways to contribute:
- Fix and [report bugs](https://github.com/influxdata/telegraf/issues/new)
- [Improve documentation](https://github.com/influxdata/telegraf/issues?q=is%3Aopen+label%3Adocumentation)
- [Review code and feature proposals](https://github.com/influxdata/telegraf/pulls)
- Answer questions on github and on the [Community Site](https://community.influxdata.com/)
- [Contribute plugins](CONTRIBUTING.md)

## Installation:

You can download the binaries directly from the [downloads](https://www.influxdata.com/downloads) page
or from the [releases](https://github.com/influxdata/telegraf/releases) section.

### Ansible Role:

Ansible role: https://github.com/rossmcdonald/telegraf

### From Source:

Telegraf requires golang version 1.8+, the Makefile requires GNU make.

Dependencies are managed with [gdm](https://github.com/sparrc/gdm),
which is installed by the Makefile if you don't have it already.

1. [Install Go](https://golang.org/doc/install)
2. [Setup your GOPATH](https://golang.org/doc/code.html#GOPATH)
3. Run `go get -d github.com/influxdata/telegraf`
4. Run `cd $GOPATH/src/github.com/influxdata/telegraf`
5. Run `make`

### Nightly Builds

These builds are generated from the master branch:
- [telegraf_nightly_amd64.deb](https://dl.influxdata.com/telegraf/nightlies/telegraf_nightly_amd64.deb)
- [telegraf_nightly_arm64.deb](https://dl.influxdata.com/telegraf/nightlies/telegraf_nightly_arm64.deb)
- [telegraf-nightly.arm64.rpm](https://dl.influxdata.com/telegraf/nightlies/telegraf-nightly.arm64.rpm)
- [telegraf_nightly_armel.deb](https://dl.influxdata.com/telegraf/nightlies/telegraf_nightly_armel.deb)
- [telegraf-nightly.armel.rpm](https://dl.influxdata.com/telegraf/nightlies/telegraf-nightly.armel.rpm)
- [telegraf_nightly_armhf.deb](https://dl.influxdata.com/telegraf/nightlies/telegraf_nightly_armhf.deb)
- [telegraf-nightly.armv6hl.rpm](https://dl.influxdata.com/telegraf/nightlies/telegraf-nightly.armv6hl.rpm)
- [telegraf-nightly_freebsd_amd64.tar.gz](https://dl.influxdata.com/telegraf/nightlies/telegraf-nightly_freebsd_amd64.tar.gz)
- [telegraf-nightly_freebsd_i386.tar.gz](https://dl.influxdata.com/telegraf/nightlies/telegraf-nightly_freebsd_i386.tar.gz)
- [telegraf_nightly_i386.deb](https://dl.influxdata.com/telegraf/nightlies/telegraf_nightly_i386.deb)
- [telegraf-nightly.i386.rpm](https://dl.influxdata.com/telegraf/nightlies/telegraf-nightly.i386.rpm)
- [telegraf-nightly_linux_amd64.tar.gz](https://dl.influxdata.com/telegraf/nightlies/telegraf-nightly_linux_amd64.tar.gz)
- [telegraf-nightly_linux_arm64.tar.gz](https://dl.influxdata.com/telegraf/nightlies/telegraf-nightly_linux_arm64.tar.gz)
- [telegraf-nightly_linux_armel.tar.gz](https://dl.influxdata.com/telegraf/nightlies/telegraf-nightly_linux_armel.tar.gz)
- [telegraf-nightly_linux_armhf.tar.gz](https://dl.influxdata.com/telegraf/nightlies/telegraf-nightly_linux_armhf.tar.gz)
- [telegraf-nightly_linux_i386.tar.gz](https://dl.influxdata.com/telegraf/nightlies/telegraf-nightly_linux_i386.tar.gz)
- [telegraf-nightly_linux_s390x.tar.gz](https://dl.influxdata.com/telegraf/nightlies/telegraf-nightly_linux_s390x.tar.gz)
- [telegraf_nightly_s390x.deb](https://dl.influxdata.com/telegraf/nightlies/telegraf_nightly_s390x.deb)
- [telegraf-nightly.s390x.rpm](https://dl.influxdata.com/telegraf/nightlies/telegraf-nightly.s390x.rpm)
- [telegraf-nightly_windows_amd64.zip](https://dl.influxdata.com/telegraf/nightlies/telegraf-nightly_windows_amd64.zip)
- [telegraf-nightly_windows_i386.zip](https://dl.influxdata.com/telegraf/nightlies/telegraf-nightly_windows_i386.zip)
- [telegraf-nightly.x86_64.rpm](https://dl.influxdata.com/telegraf/nightlies/telegraf-nightly.x86_64.rpm)
- [telegraf-static-nightly_linux_amd64.tar.gz](https://dl.influxdata.com/telegraf/nightlies/telegraf-static-nightly_linux_amd64.tar.gz)

## How to use it:

See usage with:

```
./telegraf --help
```

#### Generate a telegraf config file:

```
./telegraf config > telegraf.conf
```

#### Generate config with only cpu input & influxdb output plugins defined:

```
./telegraf --input-filter cpu --output-filter influxdb config
```

#### Run a single telegraf collection, outputing metrics to stdout:

```
./telegraf --config telegraf.conf --test
```

#### Run telegraf with all plugins defined in config file:

```
./telegraf --config telegraf.conf
```

#### Run telegraf, enabling the cpu & memory input, and influxdb output plugins:

```
./telegraf --config telegraf.conf --input-filter cpu:mem --output-filter influxdb
```


## Configuration

See the [configuration guide](docs/CONFIGURATION.md) for a rundown of the more advanced
configuration options.

## Input Plugins

* [aerospike](./plugins/inputs/aerospike)
* [amqp_consumer](./plugins/inputs/amqp_consumer) (rabbitmq)
* [apache](./plugins/inputs/apache)
* [aws cloudwatch](./plugins/inputs/cloudwatch)
* [bcache](./plugins/inputs/bcache)
* [bond](./plugins/inputs/bond)
* [cassandra](./plugins/inputs/cassandra) (deprecated, use [jolokia2](./plugins/inputs/jolokia2))
* [ceph](./plugins/inputs/ceph)
* [cgroup](./plugins/inputs/cgroup)
* [chrony](./plugins/inputs/chrony)
* [consul](./plugins/inputs/consul)
* [conntrack](./plugins/inputs/conntrack)
* [couchbase](./plugins/inputs/couchbase)
* [couchdb](./plugins/inputs/couchdb)
* [DC/OS](./plugins/inputs/dcos)
* [disque](./plugins/inputs/disque)
* [dmcache](./plugins/inputs/dmcache)
* [dns query time](./plugins/inputs/dns_query)
* [docker](./plugins/inputs/docker)
* [dovecot](./plugins/inputs/dovecot)
* [elasticsearch](./plugins/inputs/elasticsearch)
* [exec](./plugins/inputs/exec) (generic executable plugin, support JSON, influx, graphite and nagios)
* [fail2ban](./plugins/inputs/fail2ban)
* [fibaro](./plugins/inputs/fibaro)
* [filestat](./plugins/inputs/filestat)
* [fluentd](./plugins/inputs/fluentd)
* [graylog](./plugins/inputs/graylog)
* [haproxy](./plugins/inputs/haproxy)
* [hddtemp](./plugins/inputs/hddtemp)
* [http](./plugins/inputs/http) (generic HTTP plugin, supports using input data formats)
* [http_response](./plugins/inputs/http_response)
* [httpjson](./plugins/inputs/httpjson) (generic JSON-emitting http service plugin)
* [internal](./plugins/inputs/internal)
* [influxdb](./plugins/inputs/influxdb)
* [interrupts](./plugins/inputs/interrupts)
* [ipmi_sensor](./plugins/inputs/ipmi_sensor)
* [iptables](./plugins/inputs/iptables)
* [ipset](./plugins/inputs/ipset)
* [jolokia](./plugins/inputs/jolokia) (deprecated, use [jolokia2](./plugins/inputs/jolokia2))
* [jolokia2](./plugins/inputs/jolokia2) (java, cassandra, kafka)
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
* [nats](./plugins/inputs/nats)
* [net_response](./plugins/inputs/net_response)
* [nginx](./plugins/inputs/nginx)
* [nginx_plus](./plugins/inputs/nginx_plus)
* [nsq](./plugins/inputs/nsq)
* [nstat](./plugins/inputs/nstat)
* [ntpq](./plugins/inputs/ntpq)
* [nvidia_smi](./plugins/inputs/nvidia_smi)
* [openldap](./plugins/inputs/openldap)
* [opensmtpd](./plugins/inputs/opensmtpd)
* [pf](./plugins/inputs/pf)
* [phpfpm](./plugins/inputs/phpfpm)
* [phusion passenger](./plugins/inputs/passenger)
* [ping](./plugins/inputs/ping)
* [postfix](./plugins/inputs/postfix)
* [postgresql_extensible](./plugins/inputs/postgresql_extensible)
* [postgresql](./plugins/inputs/postgresql)
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
* [smart](./plugins/inputs/smart)
* [snmp](./plugins/inputs/snmp)
* [snmp_legacy](./plugins/inputs/snmp_legacy)
* [solr](./plugins/inputs/solr)
* [sql server](./plugins/inputs/sqlserver) (microsoft)
* [teamspeak](./plugins/inputs/teamspeak)
* [tomcat](./plugins/inputs/tomcat)
* [twemproxy](./plugins/inputs/twemproxy)
* [unbound](./plugins/inputs/unbound)
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
  * [papertrail](./plugins/inputs/webhooks/papertrail)
  * [particle](./plugins/inputs/webhooks/particle)
  * [rollbar](./plugins/inputs/webhooks/rollbar)
* [zipkin](./plugins/inputs/zipkin)

Telegraf is able to parse the following input data formats into metrics, these
formats may be used with input plugins supporting the `data_format` option:

* [InfluxDB Line Protocol](./docs/DATA_FORMATS_INPUT.md#influx)
* [JSON](./docs/DATA_FORMATS_INPUT.md#json)
* [Graphite](./docs/DATA_FORMATS_INPUT.md#graphite)
* [Value](./docs/DATA_FORMATS_INPUT.md#value)
* [Nagios](./docs/DATA_FORMATS_INPUT.md#nagios)
* [Collectd](./docs/DATA_FORMATS_INPUT.md#collectd)
* [Dropwizard](./docs/DATA_FORMATS_INPUT.md#dropwizard)

## Processor Plugins

* [printer](./plugins/processors/printer)
* [override](./plugins/processors/override)

## Aggregator Plugins

* [basicstats](./plugins/aggregators/basicstats)
* [minmax](./plugins/aggregators/minmax)
* [histogram](./plugins/aggregators/histogram)

## Output Plugins

* [influxdb](./plugins/outputs/influxdb)
* [amon](./plugins/outputs/amon)
* [amqp](./plugins/outputs/amqp) (rabbitmq)
* [aws kinesis](./plugins/outputs/kinesis)
* [aws cloudwatch](./plugins/outputs/cloudwatch)
* [cratedb](./plugins/outputs/cratedb)
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
* [wavefront](./plugins/outputs/wavefront)
