# Telegraf [![Circle CI](https://circleci.com/gh/influxdata/telegraf.svg?style=svg)](https://circleci.com/gh/influxdata/telegraf) [![Docker pulls](https://img.shields.io/docker/pulls/library/telegraf.svg)](https://hub.docker.com/_/telegraf/)

Telegraf is an agent for collecting, processing, aggregating, and writing metrics.

Design goals are to have a minimal memory footprint with a plugin system so
that developers in the community can easily add support for collecting
metrics.

Telegraf is plugin-driven and has the concept of 4 distinct plugin types:

1. [Input Plugins](#input-plugins) collect metrics from the system, services, or 3rd party APIs
2. [Processor Plugins](#processor-plugins) transform, decorate, and/or filter metrics
3. [Aggregator Plugins](#aggregator-plugins) create aggregate metrics (e.g. mean, min, max, quantiles, etc.)
4. [Output Plugins](#output-plugins) write metrics to various destinations

New plugins are designed to be easy to contribute, pull requests are welcomed
and we work to incorporate as many pull requests as possible.

## Try in Browser :rocket:

You can try Telegraf right in your browser in the [Telegraf playground](https://rootnroll.com/d/telegraf/).

## Contributing

There are many ways to contribute:
- Fix and [report bugs](https://github.com/influxdata/telegraf/issues/new)
- [Improve documentation](https://github.com/influxdata/telegraf/issues?q=is%3Aopen+label%3Adocumentation)
- [Review code and feature proposals](https://github.com/influxdata/telegraf/pulls)
- Answer questions and discuss here on github and on the [Community Site](https://community.influxdata.com/)
- [Contribute plugins](CONTRIBUTING.md)

## Installation:

You can download the binaries directly from the [downloads](https://www.influxdata.com/downloads) page
or from the [releases](https://github.com/influxdata/telegraf/releases) section.

### Ansible Role:

Ansible role: https://github.com/rossmcdonald/telegraf

### From Source:

Telegraf requires golang version 1.10 or newer, the Makefile requires GNU make.

1. [Install Go](https://golang.org/doc/install) >=1.10 (1.12 recommended)
2. [Install dep](https://golang.github.io/dep/docs/installation.html) ==v0.5.0
3. Download Telegraf source:
   ```
   go get -d github.com/influxdata/telegraf
   ```
4. Run make from the source directory
   ```
   cd "$HOME/go/src/github.com/influxdata/telegraf"
   make
   ```

### Changelog

View the [changelog](/CHANGELOG.md) for the latest updates and changes by
version.

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
telegraf --help
```

#### Generate a telegraf config file:

```
telegraf config > telegraf.conf
```

#### Generate config with only cpu input & influxdb output plugins defined:

```
telegraf --input-filter cpu --output-filter influxdb config
```

#### Run a single telegraf collection, outputing metrics to stdout:

```
telegraf --config telegraf.conf --test
```

#### Run telegraf with all plugins defined in config file:

```
telegraf --config telegraf.conf
```

#### Run telegraf, enabling the cpu & memory input, and influxdb output plugins:
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

* [http_listener](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/http_listener)
* [kafka_consumer](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/kafka_consumer)
* [mqtt_consumer](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/mqtt_consumer)
* [nats_consumer](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/nats_consumer)
* [nsq_consumer](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/nsq_consumer)
* [logparser](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/logparser)
* [statsd](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/statsd)
* [tail](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/tail)
* [tcp_listener](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/tcp_listener)
* [udp_listener](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/udp_listener)
* [webhooks](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/webhooks)
  * [filestack](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/webhooks/filestack)
  * [github](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/webhooks/github)
  * [mandrill](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/webhooks/mandrill)
  * [rollbar](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/webhooks/rollbar)

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
* [graylog](https://github.com/influxdata/telegraf/tree/master/plugins/outputs/graylog)
* [instrumental](https://github.com/influxdata/telegraf/tree/master/plugins/outputs/instrumental)
* [kafka](https://github.com/influxdata/telegraf/tree/master/plugins/outputs/kafka)
* [librato](https://github.com/influxdata/telegraf/tree/master/plugins/outputs/librato)
* [mqtt](https://github.com/influxdata/telegraf/tree/master/plugins/outputs/mqtt)
* [nats](https://github.com/influxdata/telegraf/tree/master/plugins/outputs/nats)
* [nsq](https://github.com/influxdata/telegraf/tree/master/plugins/outputs/nsq)
* [opentsdb](https://github.com/influxdata/telegraf/tree/master/plugins/outputs/opentsdb)
* [prometheus](https://github.com/influxdata/telegraf/tree/master/plugins/outputs/prometheus_client)
* [riemann](https://github.com/influxdata/telegraf/tree/master/plugins/outputs/riemann)
* [warp10](https://github.com/influxdata/telegraf/tree/master/plugins/outputs/warp10)

```
telegraf --config telegraf.conf --input-filter cpu:mem --output-filter influxdb
```

## Documentation

[Latest Release Documentation][release docs].

For documentation on the latest development code see the [documentation index][devel docs].

[release docs]: https://docs.influxdata.com/telegraf
[devel docs]: docs

## Input Plugins

* [activemq](./plugins/inputs/activemq)
* [aerospike](./plugins/inputs/aerospike)
* [amqp_consumer](./plugins/inputs/amqp_consumer) (rabbitmq)
* [apache](./plugins/inputs/apache)
* [aurora](./plugins/inputs/aurora)
* [aws cloudwatch](./plugins/inputs/cloudwatch)
* [bcache](./plugins/inputs/bcache)
* [beanstalkd](./plugins/inputs/beanstalkd)
* [bind](./plugins/inputs/bind)
* [bond](./plugins/inputs/bond)
* [burrow](./plugins/inputs/burrow)
* [cassandra](./plugins/inputs/cassandra) (deprecated, use [jolokia2](./plugins/inputs/jolokia2))
* [ceph](./plugins/inputs/ceph)
* [cgroup](./plugins/inputs/cgroup)
* [chrony](./plugins/inputs/chrony)
* [cisco_telemetry_gnmi](./plugins/inputs/cisco_telemetry_gnmi)
* [cisco_telemetry_mdt](./plugins/inputs/cisco_telemetry_mdt)
* [cloud_pubsub](./plugins/inputs/cloud_pubsub) Google Cloud Pub/Sub
* [cloud_pubsub_push](./plugins/inputs/cloud_pubsub_push) Google Cloud Pub/Sub push endpoint
* [conntrack](./plugins/inputs/conntrack)
* [consul](./plugins/inputs/consul)
* [couchbase](./plugins/inputs/couchbase)
* [couchdb](./plugins/inputs/couchdb)
* [cpu](./plugins/inputs/cpu)
* [DC/OS](./plugins/inputs/dcos)
* [diskio](./plugins/inputs/diskio)
* [disk](./plugins/inputs/disk)
* [disque](./plugins/inputs/disque)
* [dmcache](./plugins/inputs/dmcache)
* [dns query time](./plugins/inputs/dns_query)
* [docker](./plugins/inputs/docker)
* [docker_log](./plugins/inputs/docker_log)
* [dovecot](./plugins/inputs/dovecot)
* [ecs](./plugins/inputs/ecs) (Amazon Elastic Container Service, Fargate)
* [elasticsearch](./plugins/inputs/elasticsearch)
* [exec](./plugins/inputs/exec) (generic executable plugin, support JSON, influx, graphite and nagios)
* [fail2ban](./plugins/inputs/fail2ban)
* [fibaro](./plugins/inputs/fibaro)
* [file](./plugins/inputs/file)
* [filestat](./plugins/inputs/filestat)
* [filecount](./plugins/inputs/filecount)
* [fireboard](/plugins/inputs/fireboard)
* [fluentd](./plugins/inputs/fluentd)
* [github](./plugins/inputs/github)
* [graylog](./plugins/inputs/graylog)
* [haproxy](./plugins/inputs/haproxy)
* [hddtemp](./plugins/inputs/hddtemp)
* [httpjson](./plugins/inputs/httpjson) (generic JSON-emitting http service plugin)
* [http_listener](./plugins/inputs/influxdb_listener) (deprecated, renamed to [influxdb_listener](/plugins/inputs/influxdb_listener))
* [http_listener_v2](./plugins/inputs/http_listener_v2)
* [http](./plugins/inputs/http) (generic HTTP plugin, supports using input data formats)
* [http_response](./plugins/inputs/http_response)
* [icinga2](./plugins/inputs/icinga2)
* [influxdb](./plugins/inputs/influxdb)
* [influxdb_listener](./plugins/inputs/influxdb_listener)
* [internal](./plugins/inputs/internal)
* [interrupts](./plugins/inputs/interrupts)
* [ipmi_sensor](./plugins/inputs/ipmi_sensor)
* [ipset](./plugins/inputs/ipset)
* [iptables](./plugins/inputs/iptables)
* [ipvs](./plugins/inputs/ipvs)
* [jenkins](./plugins/inputs/jenkins)
* [jolokia2](./plugins/inputs/jolokia2) (java, cassandra, kafka)
* [jolokia](./plugins/inputs/jolokia) (deprecated, use [jolokia2](./plugins/inputs/jolokia2))
* [jti_openconfig_telemetry](./plugins/inputs/jti_openconfig_telemetry)
* [kafka_consumer](./plugins/inputs/kafka_consumer)
* [kapacitor](./plugins/inputs/kapacitor)
* [kinesis](./plugins/inputs/kinesis_consumer)
* [kernel](./plugins/inputs/kernel)
* [kernel_vmstat](./plugins/inputs/kernel_vmstat)
* [kibana](./plugins/inputs/kibana)
* [kubernetes](./plugins/inputs/kubernetes)
* [kube_inventory](./plugins/inputs/kube_inventory)
* [leofs](./plugins/inputs/leofs)
* [linux_sysctl_fs](./plugins/inputs/linux_sysctl_fs)
* [logparser](./plugins/inputs/logparser)
* [lustre2](./plugins/inputs/lustre2)
* [mailchimp](./plugins/inputs/mailchimp)
* [mcrouter](./plugins/inputs/mcrouter)
* [memcached](./plugins/inputs/memcached)
* [mem](./plugins/inputs/mem)
* [mesos](./plugins/inputs/mesos)
* [minecraft](./plugins/inputs/minecraft)
* [mongodb](./plugins/inputs/mongodb)
* [mqtt_consumer](./plugins/inputs/mqtt_consumer)
* [multifile](./plugins/inputs/multifile)
* [mysql](./plugins/inputs/mysql)
* [nats_consumer](./plugins/inputs/nats_consumer)
* [nats](./plugins/inputs/nats)
* [neptune_apex](./plugins/inputs/neptune_apex)
* [net](./plugins/inputs/net)
* [net_response](./plugins/inputs/net_response)
* [netstat](./plugins/inputs/net)
* [nginx](./plugins/inputs/nginx)
* [nginx_plus_api](./plugins/inputs/nginx_plus_api)
* [nginx_plus](./plugins/inputs/nginx_plus)
* [nginx_upstream_check](./plugins/inputs/nginx_upstream_check)
* [nginx_vts](./plugins/inputs/nginx_vts)
* [nsq_consumer](./plugins/inputs/nsq_consumer)
* [nsq](./plugins/inputs/nsq)
* [nstat](./plugins/inputs/nstat)
* [ntpq](./plugins/inputs/ntpq)
* [nvidia_smi](./plugins/inputs/nvidia_smi)
* [openldap](./plugins/inputs/openldap)
* [opensmtpd](./plugins/inputs/opensmtpd)
* [openweathermap](./plugins/inputs/openweathermap)
* [pf](./plugins/inputs/pf)
* [pgbouncer](./plugins/inputs/pgbouncer)
* [phpfpm](./plugins/inputs/phpfpm)
* [phusion passenger](./plugins/inputs/passenger)
* [ping](./plugins/inputs/ping)
* [postfix](./plugins/inputs/postfix)
* [postgresql_extensible](./plugins/inputs/postgresql_extensible)
* [postgresql](./plugins/inputs/postgresql)
* [powerdns](./plugins/inputs/powerdns)
* [powerdns_recursor](./plugins/inputs/powerdns_recursor)
* [processes](./plugins/inputs/processes)
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
* [snmp_legacy](./plugins/inputs/snmp_legacy)
* [snmp](./plugins/inputs/snmp)
* [socket_listener](./plugins/inputs/socket_listener)
* [solr](./plugins/inputs/solr)
* [sql server](./plugins/inputs/sqlserver) (microsoft)
* [stackdriver](./plugins/inputs/stackdriver)
* [statsd](./plugins/inputs/statsd)
* [swap](./plugins/inputs/swap)
* [syslog](./plugins/inputs/syslog)
* [sysstat](./plugins/inputs/sysstat)
* [system](./plugins/inputs/system)
* [tail](./plugins/inputs/tail)
* [temp](./plugins/inputs/temp)
* [tcp_listener](./plugins/inputs/socket_listener)
* [teamspeak](./plugins/inputs/teamspeak)
* [tengine](./plugins/inputs/tengine)
* [tomcat](./plugins/inputs/tomcat)
* [twemproxy](./plugins/inputs/twemproxy)
* [udp_listener](./plugins/inputs/socket_listener)
* [unbound](./plugins/inputs/unbound)
* [uswgi](./plugins/inputs/uswgi)
* [varnish](./plugins/inputs/varnish)
* [vsphere](./plugins/inputs/vsphere) VMware vSphere
* [webhooks](./plugins/inputs/webhooks)
  * [filestack](./plugins/inputs/webhooks/filestack)
  * [github](./plugins/inputs/webhooks/github)
  * [mandrill](./plugins/inputs/webhooks/mandrill)
  * [papertrail](./plugins/inputs/webhooks/papertrail)
  * [particle](./plugins/inputs/webhooks/particle)
  * [rollbar](./plugins/inputs/webhooks/rollbar)
* [win_perf_counters](./plugins/inputs/win_perf_counters) (windows performance counters)
* [win_services](./plugins/inputs/win_services)
* [wireless](./plugins/inputs/wireless)
* [x509_cert](./plugins/inputs/x509_cert)
* [zfs](./plugins/inputs/zfs)
* [zipkin](./plugins/inputs/zipkin)
* [zookeeper](./plugins/inputs/zookeeper)

## Parsers

- [InfluxDB Line Protocol](/plugins/parsers/influx)
- [Collectd](/plugins/parsers/collectd)
- [CSV](/plugins/parsers/csv)
- [Dropwizard](/plugins/parsers/dropwizard)
- [FormUrlencoded](/plugins/parser/form_urlencoded)
- [Graphite](/plugins/parsers/graphite)
- [Grok](/plugins/parsers/grok)
- [JSON](/plugins/parsers/json)
- [Logfmt](/plugins/parsers/logfmt)
- [Nagios](/plugins/parsers/nagios)
- [Value](/plugins/parsers/value), ie: 45 or "booyah"
- [Wavefront](/plugins/parsers/wavefront)

## Serializers

- [InfluxDB Line Protocol](/plugins/serializers/influx)
- [JSON](/plugins/serializers/json)
- [Graphite](/plugins/serializers/graphite)
- [ServiceNow](/plugins/serializers/nowmetric)
- [SplunkMetric](/plugins/serializers/splunkmetric)
- [Carbon2](/plugins/serializers/carbon2)
- [Wavefront](/plugins/serializers/wavefront)

## Processor Plugins

* [converter](./plugins/processors/converter)
* [date](./plugins/processors/date)
* [enum](./plugins/processors/enum)
* [override](./plugins/processors/override)
* [parser](./plugins/processors/parser)
* [pivot](./plugins/processors/pivot)
* [printer](./plugins/processors/printer)
* [regex](./plugins/processors/regex)
* [rename](./plugins/processors/rename)
* [strings](./plugins/processors/strings)
* [topk](./plugins/processors/topk)
* [unpivot](./plugins/processors/unpivot)

## Aggregator Plugins

* [basicstats](./plugins/aggregators/basicstats)
* [final](./plugins/aggregators/final)
* [histogram](./plugins/aggregators/histogram)
* [minmax](./plugins/aggregators/minmax)
* [valuecounter](./plugins/aggregators/valuecounter)

## Output Plugins

* [influxdb](./plugins/outputs/influxdb) (InfluxDB 1.x)
* [influxdb_v2](./plugins/outputs/influxdb_v2) ([InfluxDB 2.x](https://github.com/influxdata/influxdb))
* [amon](./plugins/outputs/amon)
* [amqp](./plugins/outputs/amqp) (rabbitmq)
* [application_insights](./plugins/outputs/application_insights)
* [aws kinesis](./plugins/outputs/kinesis)
* [aws cloudwatch](./plugins/outputs/cloudwatch)
* [azure_monitor](./plugins/outputs/azure_monitor)
* [cloud_pubsub](./plugins/outputs/cloud_pubsub) Google Cloud Pub/Sub
* [cratedb](./plugins/outputs/cratedb)
* [datadog](./plugins/outputs/datadog)
* [discard](./plugins/outputs/discard)
* [elasticsearch](./plugins/outputs/elasticsearch)
* [file](./plugins/outputs/file)
* [graphite](./plugins/outputs/graphite)
* [graylog](./plugins/outputs/graylog)
* [health](./plugins/outputs/health)
* [http](./plugins/outputs/http)
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
* [stackdriver](./plugins/outputs/stackdriver)
* [syslog](./plugins/outputs/syslog)
* [tcp](./plugins/outputs/socket_writer)
* [udp](./plugins/outputs/socket_writer)
* [wavefront](./plugins/outputs/wavefront)
