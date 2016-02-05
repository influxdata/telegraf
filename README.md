# Telegraf [![Circle CI](https://circleci.com/gh/influxdata/telegraf.svg?style=svg)](https://circleci.com/gh/influxdata/telegraf)

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

NOTE: Telegraf 0.10.x is **not** backwards-compatible with previous versions
of telegraf, both in the database layout and the configuration file. 0.2.x
will continue to be supported, see below for download links.

For more details on the differences between Telegraf 0.2.x and 0.10.x, see
the [release blog post](https://influxdata.com/blog/announcing-telegraf-0-10-0/).

### Linux deb and rpm Packages:

Latest:
* http://get.influxdb.org/telegraf/telegraf_0.10.2-1_amd64.deb
* http://get.influxdb.org/telegraf/telegraf-0.10.2-1.x86_64.rpm

0.2.x:
* http://get.influxdb.org/telegraf/telegraf_0.2.4_amd64.deb
* http://get.influxdb.org/telegraf/telegraf-0.2.4-1.x86_64.rpm

##### Package Instructions:

* Telegraf binary is installed in `/usr/bin/telegraf`
* Telegraf daemon configuration file is in `/etc/telegraf/telegraf.conf`
* On sysv systems, the telegraf daemon can be controlled via
`service telegraf [action]`
* On systemd systems (such as Ubuntu 15+), the telegraf daemon can be
controlled via `systemctl [action] telegraf`

### yum/apt Repositories:

There is a yum/apt repo available for the whole InfluxData stack, see
[here](https://docs.influxdata.com/influxdb/v0.9/introduction/installation/#installation)
for instructions, replacing the `influxdb` package name with `telegraf`.

### Linux tarballs:

Latest:
* http://get.influxdb.org/telegraf/telegraf-0.10.2-1_linux_amd64.tar.gz
* http://get.influxdb.org/telegraf/telegraf-0.10.2-1_linux_i386.tar.gz
* http://get.influxdb.org/telegraf/telegraf-0.10.2-1_linux_arm.tar.gz

0.2.x:
* http://get.influxdb.org/telegraf/telegraf_linux_amd64_0.2.4.tar.gz
* http://get.influxdb.org/telegraf/telegraf_linux_386_0.2.4.tar.gz
* http://get.influxdb.org/telegraf/telegraf_linux_arm_0.2.4.tar.gz

##### tarball Instructions:

To install the full directory structure with config file, run:

```
sudo tar -C / -xvf ./telegraf-0.10.2-1_linux_amd64.tar.gz
```

To extract only the binary, run:

```
tar -zxvf telegraf-0.10.2-1_linux_amd64.tar.gz --strip-components=3 ./usr/bin/telegraf
```

### Ansible Role:

Ansible role: https://github.com/rossmcdonald/telegraf

### OSX via Homebrew:

```
brew update
brew install telegraf
```

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

See the [configuration guide](CONFIGURATION.md) for a rundown of the more advanced
configuration options.

## Supported Input Plugins

Telegraf currently has support for collecting metrics from many sources. For
more information on each, please look at the directory of the same name in
`plugins/inputs`.

Currently implemented sources:

* aerospike
* apache
* bcache
* couchdb
* disque
* docker
* elasticsearch
* exec (generic executable plugin, support JSON, influx and graphite)
* haproxy
* httpjson (generic JSON-emitting http service plugin)
* influxdb
* jolokia
* leofs
* lustre2
* mailchimp
* memcached
* mongodb
* mysql
* nginx
* nsq
* phpfpm
* phusion passenger
* ping
* postgresql
* powerdns
* procstat
* prometheus
* puppetagent
* rabbitmq
* redis
* rethinkdb
* sql server (microsoft)
* twemproxy
* zfs
* zookeeper
* sensors
* snmp
* win_perf_counters (windows performance counters)
* system
    * cpu
    * mem
    * net
    * netstat
    * disk
    * diskio
    * swap

Telegraf can also collect metrics via the following service plugins:

* statsd
* kafka_consumer
* github_webhooks

We'll be adding support for many more over the coming months. Read on if you
want to add support for another service or third-party API.

## Supported Output Plugins

* influxdb
* amon
* amqp
* aws kinesis
* aws cloudwatch
* datadog
* graphite
* kafka
* librato
* mqtt
* nsq
* opentsdb
* prometheus
* riemann

## Contributing

Please see the
[contributing guide](CONTRIBUTING.md)
for details on contributing a plugin to Telegraf.
