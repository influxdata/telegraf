# Telegraf - A native agent for InfluxDB [![Circle CI](https://circleci.com/gh/influxdata/telegraf.svg?style=svg)](https://circleci.com/gh/influxdata/telegraf)

Telegraf is an agent written in Go for collecting metrics from the system it's
running on, or from other services, and writing them into InfluxDB.

Design goals are to have a minimal memory footprint with a plugin system so
that developers in the community can easily add support for collecting metrics
from well known services (like Hadoop, Postgres, or Redis) and third party
APIs (like Mailchimp, AWS CloudWatch, or Google Analytics).

We'll eagerly accept pull requests for new plugins and will manage the set of
plugins that Telegraf supports. See the
[contributing guide](CONTRIBUTING.md) for instructions on
writing new plugins.

## Installation:

### Linux deb and rpm packages:

Latest:
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

### Version 0.3.0 Beta

Version 0.3.0 will introduce many new breaking changes to Telegraf. For starters,
plugin measurements will be aggregated into fields. This means that there will no
longer be a `cpu_usage_idle` measurement, there will be a `cpu` measurement with
a `usage_idle` field.

There will also be config file changes, meaning that your 0.2.x Telegraf config
files will no longer work properly. It is recommended that you use the
`-sample-config` flag to generate a new config file to see what the changes are.
You can also read the
[0.3.0 configuration guide](https://github.com/influxdb/telegraf/blob/0.3.0/CONFIGURATION.md)
to see some of the new features and options available.

You can read more about the justifications for the aggregated measurements
[here](https://github.com/influxdb/telegraf/issues/152), and a more detailed
breakdown of the work [here](https://github.com/influxdb/telegraf/pull/437).
Once we're closer to a full release, there will be a detailed blog post
explaining all the changes.

* http://get.influxdb.org/telegraf/telegraf_0.3.0-beta2_amd64.deb
* http://get.influxdb.org/telegraf/telegraf-0.3.0_beta2-1.x86_64.rpm
* http://get.influxdb.org/telegraf/telegraf_linux_amd64_0.3.0-beta2.tar.gz
* http://get.influxdb.org/telegraf/telegraf_linux_386_0.3.0-beta2.tar.gz
* http://get.influxdb.org/telegraf/telegraf_linux_arm_0.3.0-beta2.tar.gz

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
* Or run `telegraf -sample-config -filter cpu:mem -outputfilter influxdb > telegraf.conf`.
to create a config file with only CPU and memory plugins defined, and InfluxDB
output defined.
* Edit the configuration to match your needs.
* Run `telegraf -config telegraf.conf -test` to output one full measurement
sample to STDOUT. NOTE: you may want to run as the telegraf user if you are using
the linux packages `sudo -u telegraf telegraf -config telegraf.conf -test`
* Run `telegraf -config telegraf.conf` to gather and send metrics to configured outputs.
* Run `telegraf -config telegraf.conf -filter system:swap`.
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

## Plugin Options

There are 5 configuration options that are configurable per plugin:

* **pass**: An array of strings that is used to filter metrics generated by the
current plugin. Each string in the array is tested as a glob match against metric names
and if it matches, the metric is emitted.
* **drop**: The inverse of pass, if a metric name matches, it is not emitted.
* **tagpass**: tag names and arrays of strings that are used to filter metrics by the current plugin. Each string in the array is tested as a glob match against
the tag name, and if it matches the metric is emitted.
* **tagdrop**: The inverse of tagpass. If a tag matches, the metric is not emitted.
This is tested on metrics that have passed the tagpass test.
* **interval**: How often to gather this metric. Normal plugins use a single
global interval, but if one particular plugin should be run less or more often,
you can configure that here.

### Plugin Configuration Examples

This is a full working config that will output CPU data to an InfluxDB instance
at 192.168.59.103:8086, tagging measurements with dc="denver-1". It will output
measurements at a 10s interval and will collect per-cpu data, dropping any
measurements which begin with `cpu_time`.

```toml
[tags]
  dc = "denver-1"

[agent]
  interval = "10s"

# OUTPUTS
[outputs]
[[outputs.influxdb]]
  url = "http://192.168.59.103:8086" # required.
  database = "telegraf" # required.
  precision = "s"

# PLUGINS
[plugins]
[[plugins.cpu]]
  percpu = true
  totalcpu = false
  drop = ["cpu_time*"]
```

Below is how to configure `tagpass` and `tagdrop` parameters

```toml
[plugins]
[[plugins.cpu]]
  percpu = true
  totalcpu = false
  drop = ["cpu_time"]
  # Don't collect CPU data for cpu6 & cpu7
  [plugins.cpu.tagdrop]
    cpu = [ "cpu6", "cpu7" ]

[[plugins.disk]]
  [plugins.disk.tagpass]
    # tagpass conditions are OR, not AND.
    # If the (filesystem is ext4 or xfs) OR (the path is /opt or /home)
    # then the metric passes
    fstype = [ "ext4", "xfs" ]
    # Globs can also be used on the tag values
    path = [ "/opt", "/home*" ]
```

Below is how to configure `pass` and `drop` parameters

```toml
# Drop all metrics for guest CPU usage
[[plugins.cpu]]
  drop = [ "cpu_usage_guest" ]

# Only store inode related metrics for disks
[[plugins.disk]]
  pass = [ "disk_inodes*" ]
```


Additional plugins (or outputs) of the same type can be specified,
just define more instances in the config file:

```toml
[[plugins.cpu]]
  percpu = false
  totalcpu = true

[[plugins.cpu]]
  percpu = true
  totalcpu = false
  drop = ["cpu_time*"]
```

## Supported Plugins

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
* jolokia (remote JMX with JSON over HTTP)
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
    * io
    * net
    * netstat
    * disk
    * swap

## Supported Service Plugins

Telegraf can collect metrics via the following services:

* statsd
* kafka_consumer

We'll be adding support for many more over the coming months. Read on if you
want to add support for another service or third-party API.

## Output options

Telegraf also supports specifying multiple output sinks to send data to,
configuring each output sink is different, but examples can be
found by running `telegraf -sample-config`.

Outputs also support the same configurable options as plugins
(pass, drop, tagpass, tagdrop), added in 0.2.4

```toml
[[outputs.influxdb]]
  urls = [ "http://localhost:8086" ]
  database = "telegraf"
  precision = "s"
  # Drop all measurements that start with "aerospike"
  drop = ["aerospike*"]

[[outputs.influxdb]]
  urls = [ "http://localhost:8086" ]
  database = "telegraf-aerospike-data"
  precision = "s"
  # Only accept aerospike data:
  pass = ["aerospike*"]

[[outputs.influxdb]]
  urls = [ "http://localhost:8086" ]
  database = "telegraf-cpu0-data"
  precision = "s"
  # Only store measurements where the tag "cpu" matches the value "cpu0"
  [outputs.influxdb.tagpass]
    cpu = ["cpu0"]
```


## Supported Outputs

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
