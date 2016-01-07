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

NOTE: Telegraf 0.3.x is **not** backwards-compatible with previous versions of
telegraf, both in the database layout and the configuration file. 0.2.x will
continue to be supported, see below for download links.

TODO: link to blog post about 0.3.x changes.

### Linux deb and rpm packages:

Latest:
* http://get.influxdb.org/telegraf/telegraf_0.3.0_amd64.deb
* http://get.influxdb.org/telegraf/telegraf-0.3.0-1.x86_64.rpm

0.2.x:
* http://get.influxdb.org/telegraf/telegraf_0.2.4_amd64.deb
* http://get.influxdb.org/telegraf/telegraf-0.2.4-1.x86_64.rpm

##### Package instructions:

* Telegraf binary is installed in `/opt/telegraf/telegraf`
* Telegraf daemon configuration file is in `/etc/opt/telegraf/telegraf.conf`
* On sysv systems, the telegraf daemon can be controlled via
`service telegraf [action]`
* On systemd systems (such as Ubuntu 15+), the telegraf daemon can be
controlled via `systemctl [action] telegraf`

### Linux binaries:

Latest:
* http://get.influxdb.org/telegraf/telegraf_linux_amd64_0.3.0.tar.gz
* http://get.influxdb.org/telegraf/telegraf_linux_386_0.3.0.tar.gz
* http://get.influxdb.org/telegraf/telegraf_linux_arm_0.3.0.tar.gz

0.2.x:
* http://get.influxdb.org/telegraf/telegraf_linux_amd64_0.2.4.tar.gz
* http://get.influxdb.org/telegraf/telegraf_linux_386_0.2.4.tar.gz
* http://get.influxdb.org/telegraf/telegraf_linux_arm_0.2.4.tar.gz

##### Binary instructions:

These are standalone binaries that can be unpacked and executed on any linux
system. They can be unpacked and renamed in a location such as
`/usr/local/bin` for convenience. A config file will need to be generated,
see "How to use it" below.

### OSX via Homebrew:

```
brew update
brew install telegraf
```

### From Source:

Telegraf manages dependencies via [gdm](https://github.com/sparrc/gdm),
which gets installed via the Makefile
if you don't have it already. You also must build with golang version 1.4+.

1. [Install Go](https://golang.org/doc/install)
2. [Setup your GOPATH](https://golang.org/doc/code.html#GOPATH)
3. Run `go get github.com/influxdb/telegraf`
4. Run `cd $GOPATH/src/github.com/influxdb/telegraf`
5. Run `make`

### How to use it:

* Run `telegraf -sample-config > telegraf.conf` to create an initial configuration.
* Or run `telegraf -sample-config -input-filter cpu:mem -output-filter influxdb > telegraf.conf`.
to create a config file with only CPU and memory plugins defined, and InfluxDB
output defined.
* Edit the configuration to match your needs.
* Run `telegraf -config telegraf.conf -test` to output one full measurement
sample to STDOUT. NOTE: you may want to run as the telegraf user if you are using
the linux packages `sudo -u telegraf telegraf -config telegraf.conf -test`
* Run `telegraf -config telegraf.conf` to gather and send metrics to configured outputs.
* Run `telegraf -config telegraf.conf -input-filter system:swap`.
to run telegraf with only the system & swap plugins defined in the config.

## Telegraf Options

Telegraf has a few options you can configure under the `agent` section of the
config.

* **hostname**: The hostname is passed as a tag. By default this will be
the value returned by `hostname` on the machine running Telegraf.
You can override that value here.
* **interval**: How often to gather metrics. Uses a simple number +
unit parser, e.g. "10s" for 10 seconds or "5m" for 5 minutes.
* **debug**: Set to true to gather and send metrics to STDOUT as well as
InfluxDB.

## Configuration

See the [configuration guide](CONFIGURATION.md) for a rundown of the more advanced
configuration options.

## Supported Input Plugins

**You can view usage instructions for each plugin by running**
`telegraf -usage <pluginname>`.

Telegraf currently has support for collecting metrics from:

* aerospike
* apache
* bcache
* disque
* elasticsearch
* exec (generic JSON-emitting executable plugin)
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
* phpfpm
* ping
* postgresql
* procstat
* prometheus
* puppetagent
* rabbitmq
* redis
* rethinkdb
* twemproxy
* zfs
* zookeeper
* system
    * cpu
    * mem
    * net
    * netstat
    * disk
    * diskio
    * swap

## Supported Input Service Plugins

Telegraf can collect metrics via the following services:

* statsd
* kafka_consumer

We'll be adding support for many more over the coming months. Read on if you
want to add support for another service or third-party API.

## Supported Output Plugins

* influxdb
* nsq
* kafka
* datadog
* opentsdb
* amqp (rabbitmq)
* mqtt
* librato
* prometheus
* amon
* riemann

## Contributing

Please see the
[contributing guide](CONTRIBUTING.md)
for details on contributing a plugin or output to Telegraf.
