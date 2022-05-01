
# Telegraf

![tiger](assets/TelegrafTiger.png "tiger")

[![Contribute](https://img.shields.io/badge/Contribute%20To%20Telegraf-orange.svg?logo=influx&style=for-the-badge)](https://github.com/influxdata/telegraf/blob/master/CONTRIBUTING.md) [![Slack Status](https://img.shields.io/badge/slack-join_chat-white.svg?logo=slack&style=for-the-badge)](https://www.influxdata.com/slack) [![Circle CI](https://circleci.com/gh/influxdata/telegraf.svg?style=svg)](https://circleci.com/gh/influxdata/telegraf) [![GoDoc](https://godoc.org/github.com/influxdata/telegraf?status.svg)](https://godoc.org/github.com/influxdata/telegraf) [![Docker pulls](https://img.shields.io/docker/pulls/library/telegraf.svg)](https://hub.docker.com/_/telegraf/)

Telegraf is an agent for collecting, processing, aggregating, and writing metrics. Based on a
plugin system to enable developers in the community to easily add support for additional
metric collection. There are four distinct types of plugins:

1. [Input Plugins](/docs/INPUTS.md) collect metrics from the system, services, or 3rd party APIs
2. [Processor Plugins](/docs/PROCESSORS.md) transform, decorate, and/or filter metrics
3. [Aggregator Plugins](/docs/AGGREGATORS.md) create aggregate metrics (e.g. mean, min, max, quantiles, etc.)
4. [Output Plugins](/docs/OUTPUTS.md) write metrics to various destinations

New plugins are designed to be easy to contribute, pull requests are welcomed, and we work to
incorporate as many pull requests as possible. Consider looking at the
[list of external plugins](EXTERNAL_PLUGINS.md) as well.

## Minimum Requirements

Telegraf shares the same [minimum requirements][] as Go:

- Linux kernel version 2.6.23 or later
- Windows 7 or later
- FreeBSD 11.2 or later
- MacOS 10.11 El Capitan or later

[minimum requirements]: https://github.com/golang/go/wiki/MinimumRequirements#minimum-requirements

## Obtaining Telegraf

View the [changelog](/CHANGELOG.md) for the latest updates and changes by version.

### Binary Downloads

Binary downloads are available from the [InfluxData downloads](https://www.influxdata.com/downloads)
page or from each [GitHub Releases](https://github.com/influxdata/telegraf/releases) page.

### Package Repository

InfluxData also provides a package repo that contains both DEB and RPM downloads.

For deb-based platforms (e.g. Ubuntu and Debian) run the following to add the
repo key and setup a new sources.list entry:

```shell
wget -qO- https://repos.influxdata.com/influxdb.key | sudo tee /etc/apt/trusted.gpg.d/influxdata.asc >/dev/null
echo "deb https://repos.influxdata.com/debian stable main" | sudo tee /etc/apt/sources.list.d/influxdata.list
sudo apt-get update && sudo apt-get install telegraf
```

For RPM-based platforms (e.g. RHEL, CentOS) use the following to create a repo
file and install telegraf:

```shell
cat <<EOF | sudo tee /etc/yum.repos.d/influxdata.repo
[influxdata]
name = InfluxData Repository - Stable
baseurl = https://repos.influxdata.com/stable/\$basearch/main
enabled = 1
gpgcheck = 1
gpgkey = https://repos.influxdata.com/influxdb.key
EOF
sudo yum install telegraf
```

### Build From Source

Telegraf requires Go version 1.18 or newer, the Makefile requires GNU make.

1. [Install Go](https://golang.org/doc/install) >=1.18 (1.18.0 recommended)
2. Clone the Telegraf repository:

   ```shell
   git clone https://github.com/influxdata/telegraf.git
   ```

3. Run `make` from the source directory

   ```shell
   cd telegraf
   make
   ```

### Nightly Builds

[Nightly](/docs/NIGHTLIES.md) builds are available, generated from the master branch.

### 3rd Party Builds

Builds for other platforms or package formats are provided by members of theTelegraf community.
These packages are not built, tested, or supported by the Telegraf project or InfluxData. Please
get in touch with the package author if support is needed:

- [Ansible Role](https://github.com/rossmcdonald/telegraf)
- [Chocolatey](https://chocolatey.org/packages/telegraf) by [ripclawffb](https://chocolatey.org/profiles/ripclawffb)
- [Scoop](https://github.com/ScoopInstaller/Main/blob/master/bucket/telegraf.json)
- [Snap](https://snapcraft.io/telegraf) by Laurent Sesquès (sajoupa)

## Getting Started

See usage with:

```shell
telegraf --help
```

### Generate a telegraf config file

```shell
telegraf config > telegraf.conf
```

### Generate config with only cpu input & influxdb output plugins defined

```shell
telegraf --section-filter agent:inputs:outputs --input-filter cpu --output-filter influxdb config
```

### Run a single telegraf collection, outputting metrics to stdout

```shell
telegraf --config telegraf.conf --test
```

### Run telegraf with all plugins defined in config file

```shell
telegraf --config telegraf.conf
```

### Run telegraf, enabling the cpu & memory input, and influxdb output plugins

```shell
telegraf --config telegraf.conf --input-filter cpu:mem --output-filter influxdb
```

## Contribute to the Project

Telegraf is an MIT licensed open source project and we love our community. The fastest way to get something fixed is to open a PR. Check out our [contributing guide](CONTRIBUTING.md) if you're interested in helping out. Also, join us on our [Community Slack](https://influxdata.com/slack) or [Community Page](https://community.influxdata.com/) if you have questions or comments for our engineering teams.

If your completely new to Telegraf and InfluxDB, you can also enroll for free at [InfluxDB university](https://www.influxdata.com/university/) to take courses to learn more.

## Documentation

[Latest Release Documentation](https://docs.influxdata.com/telegraf/latest/)

For documentation on the latest development code see the [documentation index](/docs).

[release docs]: https://docs.influxdata.com/telegraf
[devel docs]: docs

## Input Plugins

* [activemq](./plugins/inputs/activemq)
* [aerospike](./plugins/inputs/aerospike)
* [amqp_consumer](./plugins/inputs/amqp_consumer) (rabbitmq)
* [apache](./plugins/inputs/apache)
* [apcupsd](./plugins/inputs/apcupsd)
* [aurora](./plugins/inputs/aurora)
* [aws cloudwatch](./plugins/inputs/cloudwatch) (Amazon Cloudwatch)
* [azure_storage_queue](./plugins/inputs/azure_storage_queue)
* [bcache](./plugins/inputs/bcache)
* [beanstalkd](./plugins/inputs/beanstalkd)
* [bind](./plugins/inputs/bind)
* [bond](./plugins/inputs/bond)
* [burrow](./plugins/inputs/burrow)
* [cassandra](./plugins/inputs/cassandra) (deprecated, use [jolokia2](./plugins/inputs/jolokia2))
* [ceph](./plugins/inputs/ceph)
* [cgroup](./plugins/inputs/cgroup)
* [chrony](./plugins/inputs/chrony)
* [cisco_telemetry_gnmi](./plugins/inputs/cisco_telemetry_gnmi) (deprecated, renamed to [gnmi](/plugins/inputs/gnmi))
* [cisco_telemetry_mdt](./plugins/inputs/cisco_telemetry_mdt)
* [clickhouse](./plugins/inputs/clickhouse)
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
* [aws ecs](./plugins/inputs/ecs) (Amazon Elastic Container Service, Fargate)
* [elasticsearch](./plugins/inputs/elasticsearch)
* [ethtool](./plugins/inputs/ethtool)
* [eventhub_consumer](./plugins/inputs/eventhub_consumer) (Azure Event Hubs \& Azure IoT Hub)
* [exec](./plugins/inputs/exec) (generic executable plugin, support JSON, influx, graphite and nagios)
* [execd](./plugins/inputs/execd) (generic executable "daemon" processes)
* [fail2ban](./plugins/inputs/fail2ban)
* [fibaro](./plugins/inputs/fibaro)
* [file](./plugins/inputs/file)
* [filestat](./plugins/inputs/filestat)
* [filecount](./plugins/inputs/filecount)
* [fireboard](/plugins/inputs/fireboard)
* [fluentd](./plugins/inputs/fluentd)
* [github](./plugins/inputs/github)
* [gnmi](./plugins/inputs/gnmi)
* [graylog](./plugins/inputs/graylog)
* [haproxy](./plugins/inputs/haproxy)
* [hddtemp](./plugins/inputs/hddtemp)
* [httpjson](./plugins/inputs/httpjson) (generic JSON-emitting http service plugin)
* [http_listener](./plugins/inputs/influxdb_listener) (deprecated, renamed to [influxdb_listener](/plugins/inputs/influxdb_listener))
* [http_listener_v2](./plugins/inputs/http_listener_v2)
* [http](./plugins/inputs/http) (generic HTTP plugin, supports using input data formats)
* [http_response](./plugins/inputs/http_response)
* [icinga2](./plugins/inputs/icinga2)
* [infiniband](./plugins/inputs/infiniband)
* [influxdb](./plugins/inputs/influxdb)
* [influxdb_listener](./plugins/inputs/influxdb_listener)
* [influxdb_v2_listener](./plugins/inputs/influxdb_v2_listener)
* [intel_powerstat](plugins/inputs/intel_powerstat)
* [intel_rdt](./plugins/inputs/intel_rdt)
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
* [aws kinesis](./plugins/inputs/kinesis_consumer) (Amazon Kinesis)
* [kernel](./plugins/inputs/kernel)
* [kernel_vmstat](./plugins/inputs/kernel_vmstat)
* [kibana](./plugins/inputs/kibana)
* [kubernetes](./plugins/inputs/kubernetes)
* [kube_inventory](./plugins/inputs/kube_inventory)
* [lanz](./plugins/inputs/lanz)
* [leofs](./plugins/inputs/leofs)
* [linux_sysctl_fs](./plugins/inputs/linux_sysctl_fs)
* [logparser](./plugins/inputs/logparser) (deprecated, use [tail](/plugins/inputs/tail))
* [logstash](./plugins/inputs/logstash)
* [lustre2](./plugins/inputs/lustre2)
* [mailchimp](./plugins/inputs/mailchimp)
* [marklogic](./plugins/inputs/marklogic)
* [mcrouter](./plugins/inputs/mcrouter)
* [memcached](./plugins/inputs/memcached)
* [mem](./plugins/inputs/mem)
* [mesos](./plugins/inputs/mesos)
* [minecraft](./plugins/inputs/minecraft)
* [modbus](./plugins/inputs/modbus)
* [mongodb](./plugins/inputs/mongodb)
* [monit](./plugins/inputs/monit)
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
* [nginx_sts](./plugins/inputs/nginx_sts)
* [nginx_upstream_check](./plugins/inputs/nginx_upstream_check)
* [nginx_vts](./plugins/inputs/nginx_vts)
* [nsd](./plugins/inputs/nsd)
* [nsq_consumer](./plugins/inputs/nsq_consumer)
* [nsq](./plugins/inputs/nsq)
* [nstat](./plugins/inputs/nstat)
* [ntpq](./plugins/inputs/ntpq)
* [nvidia_smi](./plugins/inputs/nvidia_smi)
* [opcua](./plugins/inputs/opcua)
* [openldap](./plugins/inputs/openldap)
* [openntpd](./plugins/inputs/openntpd)
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
* [proxmox](./plugins/inputs/proxmox)
* [puppetagent](./plugins/inputs/puppetagent)
* [rabbitmq](./plugins/inputs/rabbitmq)
* [raindrops](./plugins/inputs/raindrops)
* [ras](./plugins/inputs/ras)
* [redfish](./plugins/inputs/redfish)
* [redis](./plugins/inputs/redis)
* [rethinkdb](./plugins/inputs/rethinkdb)
* [riak](./plugins/inputs/riak)
* [salesforce](./plugins/inputs/salesforce)
* [sensors](./plugins/inputs/sensors)
* [sflow](./plugins/inputs/sflow)
* [smart](./plugins/inputs/smart)
* [snmp_legacy](./plugins/inputs/snmp_legacy)
* [snmp](./plugins/inputs/snmp)
* [snmp_trap](./plugins/inputs/snmp_trap)
* [socket_listener](./plugins/inputs/socket_listener)
* [solr](./plugins/inputs/solr)
* [sql server](./plugins/inputs/sqlserver) (microsoft)
* [stackdriver](./plugins/inputs/stackdriver) (Google Cloud Monitoring)
* [statsd](./plugins/inputs/statsd)
* [suricata](./plugins/inputs/suricata)
* [swap](./plugins/inputs/swap)
* [synproxy](./plugins/inputs/synproxy)
* [syslog](./plugins/inputs/syslog)
* [sysstat](./plugins/inputs/sysstat)
* [systemd_units](./plugins/inputs/systemd_units)
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
* [uwsgi](./plugins/inputs/uwsgi)
* [varnish](./plugins/inputs/varnish)
* [vsphere](./plugins/inputs/vsphere) VMware vSphere
* [webhooks](./plugins/inputs/webhooks)
  * [filestack](./plugins/inputs/webhooks/filestack)
  * [github](./plugins/inputs/webhooks/github)
  * [mandrill](./plugins/inputs/webhooks/mandrill)
  * [papertrail](./plugins/inputs/webhooks/papertrail)
  * [particle](./plugins/inputs/webhooks/particle)
  * [rollbar](./plugins/inputs/webhooks/rollbar)
* [win_eventlog](./plugins/inputs/win_eventlog)
* [win_perf_counters](./plugins/inputs/win_perf_counters) (windows performance counters)
* [win_services](./plugins/inputs/win_services)
* [wireguard](./plugins/inputs/wireguard)
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

* [clone](/plugins/processors/clone)
* [converter](/plugins/processors/converter)
* [date](/plugins/processors/date)
* [dedup](/plugins/processors/dedup)
* [defaults](/plugins/processors/defaults)
* [enum](/plugins/processors/enum)
* [execd](/plugins/processors/execd)
* [ifname](/plugins/processors/ifname)
* [filepath](/plugins/processors/filepath)
* [override](/plugins/processors/override)
* [parser](/plugins/processors/parser)
* [pivot](/plugins/processors/pivot)
* [port_name](/plugins/processors/port_name)
* [printer](/plugins/processors/printer)
* [regex](/plugins/processors/regex)
* [rename](/plugins/processors/rename)
* [reverse_dns](/plugins/processors/reverse_dns)
* [s2geo](/plugins/processors/s2geo)
* [starlark](/plugins/processors/starlark)
* [strings](/plugins/processors/strings)
* [tag_limit](/plugins/processors/tag_limit)
* [template](/plugins/processors/template)
* [topk](/plugins/processors/topk)
* [unpivot](/plugins/processors/unpivot)

## Aggregator Plugins

* [basicstats](./plugins/aggregators/basicstats)
* [final](./plugins/aggregators/final)
* [histogram](./plugins/aggregators/histogram)
* [merge](./plugins/aggregators/merge)
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
* [dynatrace](./plugins/outputs/dynatrace)
* [elasticsearch](./plugins/outputs/elasticsearch)
* [exec](./plugins/outputs/exec)
* [execd](./plugins/outputs/execd)
* [file](./plugins/outputs/file)
* [graphite](./plugins/outputs/graphite)
* [graylog](./plugins/outputs/graylog)
* [health](./plugins/outputs/health)
* [http](./plugins/outputs/http)
* [instrumental](./plugins/outputs/instrumental)
* [kafka](./plugins/outputs/kafka)
* [librato](./plugins/outputs/librato)
* [logz.io](./plugins/outputs/logzio)
* [mqtt](./plugins/outputs/mqtt)
* [nats](./plugins/outputs/nats)
* [newrelic](./plugins/outputs/newrelic)
* [nsq](./plugins/outputs/nsq)
* [opentsdb](./plugins/outputs/opentsdb)
* [prometheus](./plugins/outputs/prometheus_client)
* [redistimeseries](./plugins/outputs/redistimeseries)
* [riemann](./plugins/outputs/riemann)
* [riemann_legacy](./plugins/outputs/riemann_legacy)
* [signalfx](./plugins/outputs/signalfx)
* [socket_writer](./plugins/outputs/socket_writer)
* [stackdriver](./plugins/outputs/stackdriver) (Google Cloud Monitoring)
* [syslog](./plugins/outputs/syslog)
* [tcp](./plugins/outputs/socket_writer)
* [udp](./plugins/outputs/socket_writer)
* [warp10](./plugins/outputs/warp10)
* [wavefront](./plugins/outputs/wavefront)
* [sumologic](./plugins/outputs/sumologic)
* [yandex_cloud_monitoring](./plugins/outputs/yandex_cloud_monitoring)
- [Input Plugins](/docs/INPUTS.md)
- [Output Plugins](/docs/OUTPUTS.md)
- [Processor Plugins](/docs/PROCESSORS.md)
- [Aggregator Plugins](/docs/AGGREGATORS.md)
