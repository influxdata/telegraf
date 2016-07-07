# Telegraf [![Circle CI](https://circleci.com/gh/influxdata/telegraf.svg?style=svg)](https://circleci.com/gh/influxdata/telegraf) [![Docker pulls](https://img.shields.io/docker/pulls/library/telegraf.svg)](https://hub.docker.com/_/telegraf/)

Telegraf is an agent written in Go for collecting metrics from the system it's
running on, or from other services, and writing them into InfluxDB or other
[outputs](https://github.com/influxdata/telegraf#supported-output-plugins).

Design goals are to have a minimal memory footprint with a plugin system so
that developers in the community can easily add support for collecting metrics
from well known services (like Hadoop, Postgres, or Redis) and third party
APIs (like Mailchimp, AWS CloudWatch, or Google Analytics).

New input and output plugins are designed to be easy to contribute,
we'll eagerly accept pull
requests and will manage the set of plugins that Telegraf supports.
See the [contributing guide](CONTRIBUTING.md) for instructions on writing
new plugins.

## Installation:

### Linux deb and rpm Packages:

Latest:
* https://dl.influxdata.com/telegraf/releases/telegraf_0.13.1_amd64.deb
* https://dl.influxdata.com/telegraf/releases/telegraf-0.13.1.x86_64.rpm

Latest (arm):
* https://dl.influxdata.com/telegraf/releases/telegraf_0.13.1_armhf.deb
* https://dl.influxdata.com/telegraf/releases/telegraf-0.13.1.armhf.rpm

##### Package Instructions:

* Telegraf binary is installed in `/usr/bin/telegraf`
* Telegraf daemon configuration file is in `/etc/telegraf/telegraf.conf`
* On sysv systems, the telegraf daemon can be controlled via
`service telegraf [action]`
* On systemd systems (such as Ubuntu 15+), the telegraf daemon can be
controlled via `systemctl [action] telegraf`

### yum/apt Repositories:

There is a yum/apt repo available for the whole InfluxData stack, see
[here](https://docs.influxdata.com/influxdb/v0.10/introduction/installation/#installation)
for instructions on setting up the repo. Once it is configured, you will be able
to use this repo to install & update telegraf.

### Linux tarballs:

Latest:
* https://dl.influxdata.com/telegraf/releases/telegraf-0.13.1_linux_amd64.tar.gz
* https://dl.influxdata.com/telegraf/releases/telegraf-0.13.1_linux_i386.tar.gz
* https://dl.influxdata.com/telegraf/releases/telegraf-0.13.1_linux_armhf.tar.gz

### FreeBSD tarball:

Latest:
* https://dl.influxdata.com/telegraf/releases/telegraf-0.13.1_freebsd_amd64.tar.gz

### Ansible Role:

Ansible role: https://github.com/rossmcdonald/telegraf

### OSX via Homebrew:

```
brew update
brew install telegraf
```

### Windows Binaries (EXPERIMENTAL)

Latest:
* https://dl.influxdata.com/telegraf/releases/telegraf-0.13.1_windows_amd64.zip
* https://dl.influxdata.com/telegraf/releases/telegraf-0.13.1_windows_i386.zip

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

```console
$ telegraf -help
Telegraf, The plugin-driven server agent for collecting and reporting metrics.

Usage:

  telegraf <flags>

The flags are:

  -config <file>     configuration file to load
  -test              gather metrics once, print them to stdout, and exit
  -sample-config     print out full sample configuration to stdout
  -config-directory  directory containing additional *.conf files
  -input-filter      filter the input plugins to enable, separator is :
  -output-filter     filter the output plugins to enable, separator is :
  -usage             print usage for a plugin, ie, 'telegraf -usage mysql'
  -debug             print metrics as they're generated to stdout
  -quiet             run in quiet mode
  -version           print the version to stdout

Examples:

  # generate a telegraf config file:
  telegraf -sample-config > telegraf.conf

  # generate config with only cpu input & influxdb output plugins defined
  telegraf -sample-config -input-filter cpu -output-filter influxdb

  # run a single telegraf collection, outputing metrics to stdout
  telegraf -config telegraf.conf -test

  # run telegraf with all plugins defined in config file
  telegraf -config telegraf.conf

  # run telegraf, enabling the cpu & memory input, and influxdb output plugins
  telegraf -config telegraf.conf -input-filter cpu:mem -output-filter influxdb
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
* [cassandra](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/cassandra)
* [ceph](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/ceph)
* [chrony](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/chrony)
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
* [http_response](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/http_response)
* [httpjson](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/httpjson) (generic JSON-emitting http service plugin)
* [influxdb](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/influxdb)
* [ipmi_sensor](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/ipmi_sensor)
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
* [sensors ](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/sensors) (only available if built from source)
* [snmp](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/snmp)
* [sql server](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/sqlserver) (microsoft)
* [twemproxy](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/twemproxy)
* [varnish](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/varnish)
* [zfs](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/zfs)
* [zookeeper](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/zookeeper)
* [win_perf_counters ](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/win_perf_counters) (windows performance counters)
* [sysstat](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/sysstat)
* [system](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/system)
    * cpu
    * mem
    * net
    * netstat
    * disk
    * diskio
    * swap
    * processes
    * kernel (/proc/stat)

Telegraf can also collect metrics via the following service plugins:

* [statsd](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/statsd)
* [tail](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/tail)
* [udp_listener](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/udp_listener)
* [tcp_listener](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/tcp_listener)
* [mqtt_consumer](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/mqtt_consumer)
* [kafka_consumer](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/kafka_consumer)
* [nats_consumer](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/nats_consumer)
* [github_webhooks](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/github_webhooks)

We'll be adding support for many more over the coming months. Read on if you
want to add support for another service or third-party API.

## Supported Output Plugins

* [influxdb](https://github.com/influxdata/telegraf/tree/master/plugins/outputs/influxdb)
* [amon](https://github.com/influxdata/telegraf/tree/master/plugins/outputs/amon)
* [amqp](https://github.com/influxdata/telegraf/tree/master/plugins/outputs/amqp)
* [aws kinesis](https://github.com/influxdata/telegraf/tree/master/plugins/outputs/kinesis)
* [aws cloudwatch](https://github.com/influxdata/telegraf/tree/master/plugins/outputs/cloudwatch)
* [datadog](https://github.com/influxdata/telegraf/tree/master/plugins/outputs/datadog)
* [file](https://github.com/influxdata/telegraf/tree/master/plugins/outputs/file)
* [graphite](https://github.com/influxdata/telegraf/tree/master/plugins/outputs/graphite)
* [instrumental](https://github.com/influxdata/telegraf/tree/master/plugins/outputs/instrumental)
* [kafka](https://github.com/influxdata/telegraf/tree/master/plugins/outputs/kafka)
* [librato](https://github.com/influxdata/telegraf/tree/master/plugins/outputs/librato)
* [mqtt](https://github.com/influxdata/telegraf/tree/master/plugins/outputs/mqtt)
* [nsq](https://github.com/influxdata/telegraf/tree/master/plugins/outputs/nsq)
* [opentsdb](https://github.com/influxdata/telegraf/tree/master/plugins/outputs/opentsdb)
* [prometheus](https://github.com/influxdata/telegraf/tree/master/plugins/outputs/prometheus_client)
* [riemann](https://github.com/influxdata/telegraf/tree/master/plugins/outputs/riemann)

## Contributing

Please see the
[contributing guide](CONTRIBUTING.md)
for details on contributing a plugin to Telegraf.
