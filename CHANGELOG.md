## v1.2 [unreleased]

### Release Notes

### Features

- [#1564](https://github.com/influxdata/telegraf/issues/1564): Use RFC3339 timestamps in log output.
- [#1997](https://github.com/influxdata/telegraf/issues/1997): Non-default HTTP timeouts for RabbitMQ plugin.
- [#2074](https://github.com/influxdata/telegraf/pull/2074): "discard" output plugin added, primarily for testing purposes.
- [#1965](https://github.com/influxdata/telegraf/pull/1965): The JSON parser can now parse an array of objects using the same configuration.

### Bugfixes

- [#2049](https://github.com/influxdata/telegraf/pull/2049): Fix the Value data format not trimming null characters from input.
- [#1949](https://github.com/influxdata/telegraf/issues/1949): Fix windows `net` plugin.
- [#1775](https://github.com/influxdata/telegraf/issues/1775): Cache & expire metrics for delivery to prometheus

## v1.1.1 [2016-11-14]

### Bugfixes

- [#2023](https://github.com/influxdata/telegraf/issues/2023): Fix issue parsing toml durations with single quotes.

## v1.1.0 [2016-11-07]

### Release Notes

- Telegraf now supports two new types of plugins: processors & aggregators.

- On systemd Telegraf will no longer redirect it's stdout to /var/log/telegraf/telegraf.log.
On most systems, the logs will be directed to the systemd journal and can be
accessed by `journalctl -u telegraf.service`. Consult the systemd journal
documentation for configuring journald. There is also a [`logfile` config option](https://github.com/influxdata/telegraf/blob/master/etc/telegraf.conf#L70)
available in 1.1, which will allow users to easily configure telegraf to
continue sending logs to /var/log/telegraf/telegraf.log.

### Features

- [#1726](https://github.com/influxdata/telegraf/issues/1726): Processor & Aggregator plugin support.
- [#1861](https://github.com/influxdata/telegraf/pull/1861): adding the tags in the graylog output plugin
- [#1732](https://github.com/influxdata/telegraf/pull/1732): Telegraf systemd service, log to journal.
- [#1782](https://github.com/influxdata/telegraf/pull/1782): Allow numeric and non-string values for tag_keys.
- [#1694](https://github.com/influxdata/telegraf/pull/1694): Adding Gauge and Counter metric types.
- [#1606](https://github.com/influxdata/telegraf/pull/1606): Remove carraige returns from exec plugin output on Windows
- [#1674](https://github.com/influxdata/telegraf/issues/1674): elasticsearch input: configurable timeout.
- [#1607](https://github.com/influxdata/telegraf/pull/1607): Massage metric names in Instrumental output plugin
- [#1572](https://github.com/influxdata/telegraf/pull/1572): mesos improvements.
- [#1513](https://github.com/influxdata/telegraf/issues/1513): Add Ceph Cluster Performance Statistics
- [#1650](https://github.com/influxdata/telegraf/issues/1650): Ability to configure response_timeout in httpjson input.
- [#1685](https://github.com/influxdata/telegraf/issues/1685): Add additional redis metrics.
- [#1539](https://github.com/influxdata/telegraf/pull/1539): Added capability to send metrics through Http API for OpenTSDB.
- [#1471](https://github.com/influxdata/telegraf/pull/1471): iptables input plugin.
- [#1542](https://github.com/influxdata/telegraf/pull/1542): Add filestack webhook plugin.
- [#1599](https://github.com/influxdata/telegraf/pull/1599): Add server hostname for each docker measurements.
- [#1697](https://github.com/influxdata/telegraf/pull/1697): Add NATS output plugin.
- [#1407](https://github.com/influxdata/telegraf/pull/1407) & [#1915](https://github.com/influxdata/telegraf/pull/1915): HTTP service listener input plugin.
- [#1699](https://github.com/influxdata/telegraf/pull/1699): Add database blacklist option for Postgresql
- [#1791](https://github.com/influxdata/telegraf/pull/1791): Add Docker container state metrics to Docker input plugin output
- [#1755](https://github.com/influxdata/telegraf/issues/1755): Add support to SNMP for IP & MAC address conversion.
- [#1729](https://github.com/influxdata/telegraf/issues/1729): Add support to SNMP for OID index suffixes.
- [#1813](https://github.com/influxdata/telegraf/pull/1813): Change default arguments for SNMP plugin.
- [#1686](https://github.com/influxdata/telegraf/pull/1686): Mesos input plugin: very high-cardinality mesos-task metrics removed.
- [#1838](https://github.com/influxdata/telegraf/pull/1838): Logging overhaul to centralize the logger & log levels, & provide a logfile config option.
- [#1700](https://github.com/influxdata/telegraf/pull/1700): HAProxy plugin socket glob matching.
- [#1847](https://github.com/influxdata/telegraf/pull/1847): Add Kubernetes plugin for retrieving pod metrics.

### Bugfixes

- [#1955](https://github.com/influxdata/telegraf/issues/1955): Fix NATS plug-ins reconnection logic.
- [#1936](https://github.com/influxdata/telegraf/issues/1936): Set required default values in udp_listener & tcp_listener.
- [#1926](https://github.com/influxdata/telegraf/issues/1926): Fix toml unmarshal panic in Duration objects.
- [#1746](https://github.com/influxdata/telegraf/issues/1746): Fix handling of non-string values for JSON keys listed in tag_keys.
- [#1628](https://github.com/influxdata/telegraf/issues/1628): Fix mongodb input panic on version 2.2.
- [#1733](https://github.com/influxdata/telegraf/issues/1733): Fix statsd scientific notation parsing
- [#1716](https://github.com/influxdata/telegraf/issues/1716): Sensors plugin strconv.ParseFloat: parsing "": invalid syntax
- [#1530](https://github.com/influxdata/telegraf/issues/1530): Fix prometheus_client reload panic
- [#1764](https://github.com/influxdata/telegraf/issues/1764): Fix kafka consumer panic when nil error is returned down errs channel.
- [#1768](https://github.com/influxdata/telegraf/pull/1768): Speed up statsd parsing.
- [#1751](https://github.com/influxdata/telegraf/issues/1751): Fix powerdns integer parse error handling.
- [#1752](https://github.com/influxdata/telegraf/issues/1752): Fix varnish plugin defaults not being used.
- [#1517](https://github.com/influxdata/telegraf/issues/1517): Fix windows glob paths.
- [#1137](https://github.com/influxdata/telegraf/issues/1137): Fix issue loading config directory on windows.
- [#1772](https://github.com/influxdata/telegraf/pull/1772): Windows remote management interactive service fix.
- [#1702](https://github.com/influxdata/telegraf/issues/1702): sqlserver, fix issue when case sensitive collation is activated.
- [#1823](https://github.com/influxdata/telegraf/issues/1823): Fix huge allocations in http_listener when dealing with huge payloads.
- [#1833](https://github.com/influxdata/telegraf/issues/1833): Fix translating SNMP fields not in MIB.
- [#1835](https://github.com/influxdata/telegraf/issues/1835): Fix SNMP emitting empty fields.
- [#1854](https://github.com/influxdata/telegraf/pull/1853): SQL Server waitstats truncation bug.
- [#1810](https://github.com/influxdata/telegraf/issues/1810): Fix logparser common log format: numbers in ident.
- [#1793](https://github.com/influxdata/telegraf/pull/1793): Fix JSON Serialization in OpenTSDB output.
- [#1731](https://github.com/influxdata/telegraf/issues/1731): Fix Graphite template ordering, use most specific.
- [#1836](https://github.com/influxdata/telegraf/pull/1836): Fix snmp table field initialization for non-automatic table.
- [#1724](https://github.com/influxdata/telegraf/issues/1724): cgroups path being parsed as metric.
- [#1886](https://github.com/influxdata/telegraf/issues/1886): Fix phpfpm fcgi client panic when URL does not exist.
- [#1344](https://github.com/influxdata/telegraf/issues/1344): Fix config file parse error logging.
- [#1771](https://github.com/influxdata/telegraf/issues/1771): Delete nil fields in the metric maker.
- [#870](https://github.com/influxdata/telegraf/issues/870): Fix MySQL special characters in DSN parsing.
- [#1742](https://github.com/influxdata/telegraf/issues/1742): Ping input odd timeout behavior.
- [#1950](https://github.com/influxdata/telegraf/pull/1950): Switch to github.com/kballard/go-shellquote.

## v1.0.1 [2016-09-26]

### Bugfixes

- [#1775](https://github.com/influxdata/telegraf/issues/1775): Prometheus output: Fix bug with multi-batch writes.
- [#1738](https://github.com/influxdata/telegraf/issues/1738): Fix unmarshal of influxdb metrics with null tags.
- [#1773](https://github.com/influxdata/telegraf/issues/1773): Add configurable timeout to influxdb input plugin.
- [#1785](https://github.com/influxdata/telegraf/pull/1785): Fix statsd no default value panic.

## v1.0 [2016-09-08]

### Release Notes

**Breaking Change** The SNMP plugin is being deprecated in it's current form.
There is a [new SNMP plugin](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/snmp)
which fixes many of the issues and confusions
of its predecessor. For users wanting to continue to use the deprecated SNMP
plugin, you will need to change your config file from `[[inputs.snmp]]` to
`[[inputs.snmp_legacy]]`. The configuration of the new SNMP plugin is _not_
backwards-compatible.

**Breaking Change**: Aerospike main server node measurements have been renamed
aerospike_node. Aerospike namespace measurements have been renamed to
aerospike_namespace. They will also now be tagged with the node_name
that they correspond to. This has been done to differentiate measurements
that pertain to node vs. namespace statistics.

**Breaking Change**: users of github_webhooks must change to the new
`[[inputs.webhooks]]` plugin.

This means that the default github_webhooks config:

```
# A Github Webhook Event collector
[[inputs.github_webhooks]]
  ## Address and port to host Webhook listener on
  service_address = ":1618"
```

should now look like:

```
# A Webhooks Event collector
[[inputs.webhooks]]
  ## Address and port to host Webhook listener on
  service_address = ":1618"

  [inputs.webhooks.github]
    path = "/"
```

- Telegraf now supports being installed as an official windows service,
which can be installed via
`> C:\Program Files\Telegraf\telegraf.exe --service install`

- `flush_jitter` behavior has been changed. The random jitter will now be
evaluated at every flush interval, rather than once at startup. This makes it
consistent with the behavior of `collection_jitter`.

### Features

- [#1413](https://github.com/influxdata/telegraf/issues/1413): Separate container_version from container_image tag.
- [#1525](https://github.com/influxdata/telegraf/pull/1525): Support setting per-device and total metrics for Docker network and blockio.
- [#1466](https://github.com/influxdata/telegraf/pull/1466): MongoDB input plugin: adding per DB stats from db.stats()
- [#1503](https://github.com/influxdata/telegraf/pull/1503): Add tls support for certs to RabbitMQ input plugin
- [#1289](https://github.com/influxdata/telegraf/pull/1289): webhooks input plugin. Thanks @francois2metz and @cduez!
- [#1247](https://github.com/influxdata/telegraf/pull/1247): rollbar webhook plugin.
- [#1408](https://github.com/influxdata/telegraf/pull/1408): mandrill webhook plugin.
- [#1402](https://github.com/influxdata/telegraf/pull/1402): docker-machine/boot2docker no longer required for unit tests.
- [#1350](https://github.com/influxdata/telegraf/pull/1350): cgroup input plugin.
- [#1369](https://github.com/influxdata/telegraf/pull/1369): Add input plugin for consuming metrics from NSQD.
- [#1369](https://github.com/influxdata/telegraf/pull/1480): add ability to read redis from a socket.
- [#1387](https://github.com/influxdata/telegraf/pull/1387): **Breaking Change** - Redis `role` tag renamed to `replication_role` to avoid global_tags override
- [#1437](https://github.com/influxdata/telegraf/pull/1437): Fetching Galera status metrics in MySQL
- [#1500](https://github.com/influxdata/telegraf/pull/1500): Aerospike plugin refactored to use official client lib.
- [#1434](https://github.com/influxdata/telegraf/pull/1434): Add measurement name arg to logparser plugin.
- [#1479](https://github.com/influxdata/telegraf/pull/1479): logparser: change resp_code from a field to a tag.
- [#1411](https://github.com/influxdata/telegraf/pull/1411): Implement support for fetching hddtemp data
- [#1340](https://github.com/influxdata/telegraf/issues/1340): statsd: do not log every dropped metric.
- [#1368](https://github.com/influxdata/telegraf/pull/1368): Add precision rounding to all metrics on collection.
- [#1390](https://github.com/influxdata/telegraf/pull/1390): Add support for Tengine
- [#1320](https://github.com/influxdata/telegraf/pull/1320): Logparser input plugin for parsing grok-style log patterns.
- [#1397](https://github.com/influxdata/telegraf/issues/1397): ElasticSearch: now supports connecting to ElasticSearch via SSL
- [#1262](https://github.com/influxdata/telegraf/pull/1261): Add graylog input pluging.
- [#1294](https://github.com/influxdata/telegraf/pull/1294): consul input plugin. Thanks @harnash
- [#1164](https://github.com/influxdata/telegraf/pull/1164): conntrack input plugin. Thanks @robinpercy!
- [#1165](https://github.com/influxdata/telegraf/pull/1165): vmstat input plugin. Thanks @jshim-xm!
- [#1208](https://github.com/influxdata/telegraf/pull/1208): Standardized AWS credentials evaluation & wildcard CloudWatch dimensions. Thanks @johnrengelman!
- [#1264](https://github.com/influxdata/telegraf/pull/1264): Add SSL config options to http_response plugin.
- [#1272](https://github.com/influxdata/telegraf/pull/1272): graphite parser: add ability to specify multiple tag keys, for consistency with influxdb parser.
- [#1265](https://github.com/influxdata/telegraf/pull/1265): Make dns lookups for chrony configurable. Thanks @zbindenren!
- [#1275](https://github.com/influxdata/telegraf/pull/1275): Allow wildcard filtering of varnish stats.
- [#1142](https://github.com/influxdata/telegraf/pull/1142): Support for glob patterns in exec plugin commands configuration.
- [#1278](https://github.com/influxdata/telegraf/pull/1278): RabbitMQ input: made url parameter optional by using DefaultURL (http://localhost:15672) if not specified
- [#1197](https://github.com/influxdata/telegraf/pull/1197): Limit AWS GetMetricStatistics requests to 10 per second.
- [#1278](https://github.com/influxdata/telegraf/pull/1278) & [#1288](https://github.com/influxdata/telegraf/pull/1288) & [#1295](https://github.com/influxdata/telegraf/pull/1295): RabbitMQ/Apache/InfluxDB inputs: made url(s) parameter optional by using reasonable input defaults if not specified
- [#1296](https://github.com/influxdata/telegraf/issues/1296): Refactor of flush_jitter argument.
- [#1213](https://github.com/influxdata/telegraf/issues/1213): Add inactive & active memory to mem plugin.
- [#1543](https://github.com/influxdata/telegraf/pull/1543): Official Windows service.
- [#1414](https://github.com/influxdata/telegraf/pull/1414): Forking sensors command to remove C package dependency.
- [#1389](https://github.com/influxdata/telegraf/pull/1389): Add a new SNMP plugin.

### Bugfixes

- [#1619](https://github.com/influxdata/telegraf/issues/1619): Fix `make windows` build target
- [#1519](https://github.com/influxdata/telegraf/pull/1519): Fix error race conditions and partial failures.
- [#1477](https://github.com/influxdata/telegraf/issues/1477): nstat: fix inaccurate config panic.
- [#1481](https://github.com/influxdata/telegraf/issues/1481): jolokia: fix handling multiple multi-dimensional attributes.
- [#1430](https://github.com/influxdata/telegraf/issues/1430): Fix prometheus character sanitizing. Sanitize more win_perf_counters characters.
- [#1534](https://github.com/influxdata/telegraf/pull/1534): Add diskio io_time to FreeBSD & report timing metrics as ms (as linux does).
- [#1379](https://github.com/influxdata/telegraf/issues/1379): Fix covering Amazon Linux for post remove flow.
- [#1584](https://github.com/influxdata/telegraf/issues/1584): procstat missing fields: read/write bytes & count
- [#1472](https://github.com/influxdata/telegraf/pull/1472): diskio input plugin: set 'skip_serial_number = true' by default to avoid high cardinality.
- [#1426](https://github.com/influxdata/telegraf/pull/1426): nil metrics panic fix.
- [#1384](https://github.com/influxdata/telegraf/pull/1384): Fix datarace in apache input plugin.
- [#1399](https://github.com/influxdata/telegraf/issues/1399): Add `read_repairs` statistics to riak plugin.
- [#1405](https://github.com/influxdata/telegraf/issues/1405): Fix memory/connection leak in prometheus input plugin.
- [#1378](https://github.com/influxdata/telegraf/issues/1378): Trim BOM from config file for Windows support.
- [#1339](https://github.com/influxdata/telegraf/issues/1339): Prometheus client output panic on service reload.
- [#1461](https://github.com/influxdata/telegraf/pull/1461): Prometheus parser, protobuf format header fix.
- [#1334](https://github.com/influxdata/telegraf/issues/1334): Prometheus output, metric refresh and caching fixes.
- [#1432](https://github.com/influxdata/telegraf/issues/1432): Panic fix for multiple graphite outputs under very high load.
- [#1412](https://github.com/influxdata/telegraf/pull/1412): Instrumental output has better reconnect behavior
- [#1460](https://github.com/influxdata/telegraf/issues/1460): Remove PID from procstat plugin to fix cardinality issues.
- [#1427](https://github.com/influxdata/telegraf/issues/1427): Cassandra input: version 2.x "column family" fix.
- [#1463](https://github.com/influxdata/telegraf/issues/1463): Shared WaitGroup in Exec plugin
- [#1436](https://github.com/influxdata/telegraf/issues/1436): logparser: honor modifiers in "pattern" config.
- [#1418](https://github.com/influxdata/telegraf/issues/1418): logparser: error and exit on file permissions/missing errors.
- [#1499](https://github.com/influxdata/telegraf/pull/1499): Make the user able to specify full path for HAproxy stats
- [#1521](https://github.com/influxdata/telegraf/pull/1521): Fix Redis url, an extra "tcp://" was added.
- [#1330](https://github.com/influxdata/telegraf/issues/1330): Fix exec plugin panic when using single binary.
- [#1336](https://github.com/influxdata/telegraf/issues/1336): Fixed incorrect prometheus metrics source selection.
- [#1112](https://github.com/influxdata/telegraf/issues/1112): Set default Zookeeper chroot to empty string.
- [#1335](https://github.com/influxdata/telegraf/issues/1335): Fix overall ping timeout to be calculated based on per-ping timeout.
- [#1374](https://github.com/influxdata/telegraf/pull/1374): Change "default" retention policy to "".
- [#1377](https://github.com/influxdata/telegraf/issues/1377): Graphite output mangling '%' character.
- [#1396](https://github.com/influxdata/telegraf/pull/1396): Prometheus input plugin now supports x509 certs authentication
- [#1252](https://github.com/influxdata/telegraf/pull/1252) & [#1279](https://github.com/influxdata/telegraf/pull/1279): Fix systemd service. Thanks @zbindenren & @PierreF!
- [#1221](https://github.com/influxdata/telegraf/pull/1221): Fix influxdb n_shards counter.
- [#1258](https://github.com/influxdata/telegraf/pull/1258): Fix potential kernel plugin integer parse error.
- [#1268](https://github.com/influxdata/telegraf/pull/1268): Fix potential influxdb input type assertion panic.
- [#1283](https://github.com/influxdata/telegraf/pull/1283): Still send processes metrics if a process exited during metric collection.
- [#1297](https://github.com/influxdata/telegraf/issues/1297): disk plugin panic when usage grab fails.
- [#1316](https://github.com/influxdata/telegraf/pull/1316): Removed leaked "database" tag on redis metrics. Thanks @PierreF!
- [#1323](https://github.com/influxdata/telegraf/issues/1323): Processes plugin: fix potential error with /proc/net/stat directory.
- [#1322](https://github.com/influxdata/telegraf/issues/1322): Fix rare RHEL 5.2 panic in gopsutil diskio gathering function.
- [#1586](https://github.com/influxdata/telegraf/pull/1586): Remove IF NOT EXISTS from influxdb output database creation.
- [#1600](https://github.com/influxdata/telegraf/issues/1600): Fix quoting with text values in postgresql_extensible plugin.
- [#1425](https://github.com/influxdata/telegraf/issues/1425): Fix win_perf_counter "index out of range" panic.
- [#1634](https://github.com/influxdata/telegraf/issues/1634): Fix ntpq panic when field is missing.
- [#1637](https://github.com/influxdata/telegraf/issues/1637): Sanitize graphite output field names.
- [#1695](https://github.com/influxdata/telegraf/pull/1695): Fix MySQL plugin not sending 0 value fields.

## v0.13.1 [2016-05-24]

### Release Notes

- net_response and http_response plugins timeouts will now accept duration
strings, ie, "2s" or "500ms".
- Input plugin Gathers will no longer be logged by default, but a Gather for
_each_ plugin will be logged in Debug mode.
- Debug mode will no longer print every point added to the accumulator. This
functionality can be duplicated using the `file` output plugin and printing
to "stdout".

### Features

- [#1173](https://github.com/influxdata/telegraf/pull/1173): varnish input plugin. Thanks @sfox-xmatters!
- [#1138](https://github.com/influxdata/telegraf/pull/1138): nstat input plugin. Thanks @Maksadbek!
- [#1139](https://github.com/influxdata/telegraf/pull/1139): instrumental output plugin. Thanks @jasonroelofs!
- [#1172](https://github.com/influxdata/telegraf/pull/1172): Ceph storage stats. Thanks @robinpercy!
- [#1233](https://github.com/influxdata/telegraf/pull/1233): Updated golint gopsutil dependency.
- [#1238](https://github.com/influxdata/telegraf/pull/1238): chrony input plugin. Thanks @zbindenren!
- [#479](https://github.com/influxdata/telegraf/issues/479): per-plugin execution time added to debug output.
- [#1249](https://github.com/influxdata/telegraf/issues/1249): influxdb output: added write_consistency argument.

### Bugfixes

- [#1195](https://github.com/influxdata/telegraf/pull/1195): Docker panic on timeout. Thanks @zstyblik!
- [#1211](https://github.com/influxdata/telegraf/pull/1211): mongodb input. Fix possible panic. Thanks @kols!
- [#1215](https://github.com/influxdata/telegraf/pull/1215): Fix for possible gopsutil-dependent plugin hangs.
- [#1228](https://github.com/influxdata/telegraf/pull/1228): Fix service plugin host tag overwrite.
- [#1198](https://github.com/influxdata/telegraf/pull/1198): http_response: override request Host header properly
- [#1230](https://github.com/influxdata/telegraf/issues/1230): Fix Telegraf process hangup due to a single plugin hanging.
- [#1214](https://github.com/influxdata/telegraf/issues/1214): Use TCP timeout argument in net_response plugin.
- [#1243](https://github.com/influxdata/telegraf/pull/1243): Logfile not created on systemd.

## v0.13 [2016-05-11]

### Release Notes

- **Breaking change** in jolokia plugin. See
https://github.com/influxdata/telegraf/blob/master/plugins/inputs/jolokia/README.md
for updated configuration. The plugin will now support proxy mode and will make
POST requests.

- New [agent] configuration option: `metric_batch_size`. This option tells
telegraf the maximum batch size to allow to accumulate before sending a flush
to the configured outputs. `metric_buffer_limit` now refers to the absolute
maximum number of metrics that will accumulate before metrics are dropped.

- There is no longer an option to
`flush_buffer_when_full`, this is now the default and only behavior of telegraf.

- **Breaking Change**: docker plugin tags. The cont_id tag no longer exists, it
will now be a field, and be called container_id. Additionally, cont_image and
cont_name are being renamed to container_image and container_name.

- **Breaking Change**: docker plugin measurements. The `docker_cpu`, `docker_mem`,
`docker_blkio` and `docker_net` measurements are being renamed to
`docker_container_cpu`, `docker_container_mem`, `docker_container_blkio` and
`docker_container_net`. Why? Because these metrics are
specifically tracking per-container stats. The problem with per-container stats,
in some use-cases, is that if containers are short-lived AND names are not
kept consistent, then the series cardinality will balloon very quickly.
So adding "container" to each metric will:
(1) make it more clear that these metrics are per-container, and
(2) allow users to easily drop per-container metrics if cardinality is an
issue (`namedrop = ["docker_container_*"]`)

- `tagexclude` and `taginclude` are now available, which can be used to remove
tags from measurements on inputs and outputs. See
[the configuration doc](https://github.com/influxdata/telegraf/blob/master/docs/CONFIGURATION.md)
for more details.

- **Measurement filtering:** All measurement filters now match based on glob
only. Previously there was an undocumented behavior where filters would match
based on _prefix_ in addition to globs. This means that a filter like
`fielddrop = ["time_"]` will need to be changed to `fielddrop = ["time_*"]`

- **datadog**: measurement and field names will no longer have `_` replaced by `.`

- The following plugins have changed their tags to _not_ overwrite the host tag:
  - cassandra: `host -> cassandra_host`
  - disque: `host -> disque_host`
  - rethinkdb: `host -> rethinkdb_host`

- **Breaking Change**: The `win_perf_counters` input has been changed to
sanitize field names, replacing `/Sec` and `/sec` with `_persec`, as well as
spaces with underscores. This is needed because Graphite doesn't like slashes
and spaces, and was failing to accept metrics that had them.
The `/[sS]ec` -> `_persec` is just to make things clearer and uniform.

- **Breaking Change**: snmp plugin. The `host` tag of the snmp plugin has been
changed to the `snmp_host` tag.

- The `disk` input plugin can now be configured with the `HOST_MOUNT_PREFIX` environment variable.
This value is prepended to any mountpaths discovered before retrieving stats.
It is not included on the report path. This is necessary for reporting host disk stats when running from within a container.

### Features

- [#1031](https://github.com/influxdata/telegraf/pull/1031): Jolokia plugin proxy mode. Thanks @saiello!
- [#1017](https://github.com/influxdata/telegraf/pull/1017): taginclude and tagexclude arguments.
- [#1015](https://github.com/influxdata/telegraf/pull/1015): Docker plugin schema refactor.
- [#889](https://github.com/influxdata/telegraf/pull/889): Improved MySQL plugin. Thanks @maksadbek!
- [#1060](https://github.com/influxdata/telegraf/pull/1060): TTL metrics added to MongoDB input plugin
- [#1056](https://github.com/influxdata/telegraf/pull/1056): Don't allow inputs to overwrite host tags.
- [#1035](https://github.com/influxdata/telegraf/issues/1035): Add `user`, `exe`, `pidfile` tags to procstat plugin.
- [#1041](https://github.com/influxdata/telegraf/issues/1041): Add `n_cpus` field to the system plugin.
- [#1072](https://github.com/influxdata/telegraf/pull/1072): New Input Plugin: filestat.
- [#1066](https://github.com/influxdata/telegraf/pull/1066): Replication lag metrics for MongoDB input plugin
- [#1086](https://github.com/influxdata/telegraf/pull/1086): Ability to specify AWS keys in config file. Thanks @johnrengleman!
- [#1096](https://github.com/influxdata/telegraf/pull/1096): Performance refactor of running output buffers.
- [#967](https://github.com/influxdata/telegraf/issues/967): Buffer logging improvements.
- [#1107](https://github.com/influxdata/telegraf/issues/1107): Support lustre2 job stats. Thanks @hanleyja!
- [#1122](https://github.com/influxdata/telegraf/pull/1122): Support setting config path through env variable and default paths.
- [#1128](https://github.com/influxdata/telegraf/pull/1128): MongoDB jumbo chunks metric for MongoDB input plugin
- [#1146](https://github.com/influxdata/telegraf/pull/1146): HAProxy socket support. Thanks weshmashian!

### Bugfixes

- [#1050](https://github.com/influxdata/telegraf/issues/1050): jolokia plugin - do not overwrite host tag. Thanks @saiello!
- [#921](https://github.com/influxdata/telegraf/pull/921): mqtt_consumer stops gathering metrics. Thanks @chaton78!
- [#1013](https://github.com/influxdata/telegraf/pull/1013): Close dead riemann output connections. Thanks @echupriyanov!
- [#1012](https://github.com/influxdata/telegraf/pull/1012): Set default tags in test accumulator.
- [#1024](https://github.com/influxdata/telegraf/issues/1024): Don't replace `.` with `_` in datadog output.
- [#1058](https://github.com/influxdata/telegraf/issues/1058): Fix possible leaky TCP connections in influxdb output.
- [#1044](https://github.com/influxdata/telegraf/pull/1044): Fix SNMP OID possible collisions. Thanks @relip
- [#1022](https://github.com/influxdata/telegraf/issues/1022): Dont error deb/rpm install on systemd errors.
- [#1078](https://github.com/influxdata/telegraf/issues/1078): Use default AWS credential chain.
- [#1070](https://github.com/influxdata/telegraf/issues/1070): SQL Server input. Fix datatype conversion.
- [#1089](https://github.com/influxdata/telegraf/issues/1089): Fix leaky TCP connections in phpfpm plugin.
- [#914](https://github.com/influxdata/telegraf/issues/914): Telegraf can drop metrics on full buffers.
- [#1098](https://github.com/influxdata/telegraf/issues/1098): Sanitize invalid OpenTSDB characters.
- [#1110](https://github.com/influxdata/telegraf/pull/1110): Sanitize * to - in graphite serializer. Thanks @goodeggs!
- [#1118](https://github.com/influxdata/telegraf/pull/1118): Sanitize Counter names for `win_perf_counters` input.
- [#1125](https://github.com/influxdata/telegraf/pull/1125): Wrap all exec command runners with a timeout, so hung os processes don't halt Telegraf.
- [#1113](https://github.com/influxdata/telegraf/pull/1113): Set MaxRetry and RequiredAcks defaults in Kafka output.
- [#1090](https://github.com/influxdata/telegraf/issues/1090): [agent] and [global_tags] config sometimes not getting applied.
- [#1133](https://github.com/influxdata/telegraf/issues/1133): Use a timeout for docker list & stat cmds.
- [#1052](https://github.com/influxdata/telegraf/issues/1052): Docker panic fix when decode fails.
- [#1136](https://github.com/influxdata/telegraf/pull/1136): "DELAYED" Inserts were deprecated in MySQL 5.6.6. Thanks @PierreF

## v0.12.1 [2016-04-14]

### Release Notes
- Breaking change in the dovecot input plugin. See Features section below.
- Graphite output templates are now supported. See
https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md#graphite
- Possible breaking change for the librato and graphite outputs. Telegraf will
no longer insert field names when the field is simply named `value`. This is
because the `value` field is redundant in the graphite/librato context.

### Features
- [#1009](https://github.com/influxdata/telegraf/pull/1009): Cassandra input plugin. Thanks @subhachandrachandra!
- [#976](https://github.com/influxdata/telegraf/pull/976): Reduce allocations in the UDP and statsd inputs.
- [#979](https://github.com/influxdata/telegraf/pull/979): Reduce allocations in the TCP listener.
- [#992](https://github.com/influxdata/telegraf/pull/992): Refactor allocations in TCP/UDP listeners.
- [#935](https://github.com/influxdata/telegraf/pull/935): AWS Cloudwatch input plugin. Thanks @joshhardy & @ljosa!
- [#943](https://github.com/influxdata/telegraf/pull/943): http_response input plugin. Thanks @Lswith!
- [#939](https://github.com/influxdata/telegraf/pull/939): sysstat input plugin. Thanks @zbindenren!
- [#998](https://github.com/influxdata/telegraf/pull/998): **breaking change** enabled global, user and ip queries in dovecot plugin. Thanks @mikif70!
- [#1001](https://github.com/influxdata/telegraf/pull/1001): Graphite serializer templates.
- [#1008](https://github.com/influxdata/telegraf/pull/1008): Adding memstats metrics to the influxdb plugin.

### Bugfixes
- [#968](https://github.com/influxdata/telegraf/issues/968): Processes plugin gets unknown state when spaces are in (command name)
- [#969](https://github.com/influxdata/telegraf/pull/969): ipmi_sensors: allow : in password. Thanks @awaw!
- [#972](https://github.com/influxdata/telegraf/pull/972): dovecot: remove extra newline in dovecot command. Thanks @mrannanj!
- [#645](https://github.com/influxdata/telegraf/issues/645): docker plugin i/o error on closed pipe. Thanks @tripledes!

## v0.12.0 [2016-04-05]

### Features
- [#951](https://github.com/influxdata/telegraf/pull/951): Parse environment variables in the config file.
- [#948](https://github.com/influxdata/telegraf/pull/948): Cleanup config file and make default package version include all plugins (but commented).
- [#927](https://github.com/influxdata/telegraf/pull/927): Adds parsing of tags to the statsd input when using DataDog's dogstatsd extension
- [#863](https://github.com/influxdata/telegraf/pull/863): AMQP output: allow external auth. Thanks @ekini!
- [#707](https://github.com/influxdata/telegraf/pull/707): Improved prometheus plugin. Thanks @titilambert!
- [#878](https://github.com/influxdata/telegraf/pull/878): Added json serializer. Thanks @ch3lo!
- [#880](https://github.com/influxdata/telegraf/pull/880): Add the ability to specify the bearer token to the prometheus plugin. Thanks @jchauncey!
- [#882](https://github.com/influxdata/telegraf/pull/882): Fixed SQL Server Plugin issues
- [#849](https://github.com/influxdata/telegraf/issues/849): Adding ability to parse single values as an input data type.
- [#844](https://github.com/influxdata/telegraf/pull/844): postgres_extensible plugin added. Thanks @menardorama!
- [#866](https://github.com/influxdata/telegraf/pull/866): couchbase input plugin. Thanks @ljosa!
- [#789](https://github.com/influxdata/telegraf/pull/789): Support multiple field specification and `field*` in graphite templates. Thanks @chrusty!
- [#762](https://github.com/influxdata/telegraf/pull/762): Nagios parser for the exec plugin. Thanks @titilambert!
- [#848](https://github.com/influxdata/telegraf/issues/848): Provide option to omit host tag from telegraf agent.
- [#928](https://github.com/influxdata/telegraf/pull/928): Deprecating the statsd "convert_names" options, expose separator config.
- [#919](https://github.com/influxdata/telegraf/pull/919): ipmi_sensor input plugin. Thanks @ebookbug!
- [#945](https://github.com/influxdata/telegraf/pull/945): KAFKA output: codec, acks, and retry configuration. Thanks @framiere!

### Bugfixes
- [#890](https://github.com/influxdata/telegraf/issues/890): Create TLS config even if only ssl_ca is provided.
- [#884](https://github.com/influxdata/telegraf/issues/884): Do not call write method if there are 0 metrics to write.
- [#898](https://github.com/influxdata/telegraf/issues/898): Put database name in quotes, fixes special characters in the database name.
- [#656](https://github.com/influxdata/telegraf/issues/656): No longer run `lsof` on linux to get netstat data, fixes permissions issue.
- [#907](https://github.com/influxdata/telegraf/issues/907): Fix prometheus invalid label/measurement name key.
- [#841](https://github.com/influxdata/telegraf/issues/841): Fix memcached unix socket panic.
- [#873](https://github.com/influxdata/telegraf/issues/873): Fix SNMP plugin sometimes not returning metrics. Thanks @titiliambert!
- [#934](https://github.com/influxdata/telegraf/pull/934): phpfpm: Fix fcgi uri path. Thanks @rudenkovk!
- [#805](https://github.com/influxdata/telegraf/issues/805): Kafka consumer stops gathering after i/o timeout.
- [#959](https://github.com/influxdata/telegraf/pull/959): reduce mongodb & prometheus collection timeouts. Thanks @PierreF!

## v0.11.1 [2016-03-17]

### Release Notes
- Primarily this release was cut to fix [#859](https://github.com/influxdata/telegraf/issues/859)

### Features
- [#747](https://github.com/influxdata/telegraf/pull/747): Start telegraf on install & remove on uninstall. Thanks @pierref!
- [#794](https://github.com/influxdata/telegraf/pull/794): Add service reload ability. Thanks @entertainyou!

### Bugfixes
- [#852](https://github.com/influxdata/telegraf/issues/852): Windows zip package fix
- [#859](https://github.com/influxdata/telegraf/issues/859): httpjson plugin panic

## v0.11.0 [2016-03-15]

### Release Notes

### Features
- [#692](https://github.com/influxdata/telegraf/pull/770): Support InfluxDB retention policies
- [#771](https://github.com/influxdata/telegraf/pull/771): Default timeouts for input plugns. Thanks @PierreF!
- [#758](https://github.com/influxdata/telegraf/pull/758): UDP Listener input plugin, thanks @whatyouhide!
- [#769](https://github.com/influxdata/telegraf/issues/769): httpjson plugin: allow specifying SSL configuration.
- [#735](https://github.com/influxdata/telegraf/pull/735): SNMP Table feature. Thanks @titilambert!
- [#754](https://github.com/influxdata/telegraf/pull/754): docker plugin: adding `docker info` metrics to output. Thanks @titilambert!
- [#788](https://github.com/influxdata/telegraf/pull/788): -input-list and -output-list command-line options. Thanks @ebookbug!
- [#778](https://github.com/influxdata/telegraf/pull/778): Adding a TCP input listener.
- [#797](https://github.com/influxdata/telegraf/issues/797): Provide option for persistent MQTT consumer client sessions.
- [#799](https://github.com/influxdata/telegraf/pull/799): Add number of threads for procstat input plugin. Thanks @titilambert!
- [#776](https://github.com/influxdata/telegraf/pull/776): Add Zookeeper chroot option to kafka_consumer. Thanks @prune998!
- [#811](https://github.com/influxdata/telegraf/pull/811): Add processes plugin for classifying total procs on system. Thanks @titilambert!
- [#235](https://github.com/influxdata/telegraf/issues/235): Add number of users to the `system` input plugin.
- [#826](https://github.com/influxdata/telegraf/pull/826): "kernel" linux plugin for /proc/stat metrics (context switches, interrupts, etc.)
- [#847](https://github.com/influxdata/telegraf/pull/847): `ntpq`: Input plugin for running ntp query executable and gathering metrics.

### Bugfixes
- [#748](https://github.com/influxdata/telegraf/issues/748): Fix sensor plugin split on ":"
- [#722](https://github.com/influxdata/telegraf/pull/722): Librato output plugin fixes. Thanks @chrusty!
- [#745](https://github.com/influxdata/telegraf/issues/745): Fix Telegraf toml parse panic on large config files. Thanks @titilambert!
- [#781](https://github.com/influxdata/telegraf/pull/781): Fix mqtt_consumer username not being set. Thanks @chaton78!
- [#786](https://github.com/influxdata/telegraf/pull/786): Fix mqtt output username not being set. Thanks @msangoi!
- [#773](https://github.com/influxdata/telegraf/issues/773): Fix duplicate measurements in snmp plugin. Thanks @titilambert!
- [#708](https://github.com/influxdata/telegraf/issues/708): packaging: build ARM package
- [#713](https://github.com/influxdata/telegraf/issues/713): packaging: insecure permissions error on log directory
- [#816](https://github.com/influxdata/telegraf/issues/816): Fix phpfpm panic if fcgi endpoint unreachable.
- [#828](https://github.com/influxdata/telegraf/issues/828): fix net_response plugin overwriting host tag.
- [#821](https://github.com/influxdata/telegraf/issues/821): Remove postgres password from server tag. Thanks @menardorama!

## v0.10.4.1

### Release Notes
- Bug in the build script broke deb and rpm packages.

### Bugfixes
- [#750](https://github.com/influxdata/telegraf/issues/750): deb package broken
- [#752](https://github.com/influxdata/telegraf/issues/752): rpm package broken

## v0.10.4 [2016-02-24]

### Release Notes
- The pass/drop parameters have been renamed to fielddrop/fieldpass parameters,
to more accurately indicate their purpose.
- There are also now namedrop/namepass parameters for passing/dropping based
on the metric _name_.
- Experimental windows builds now available.

### Features
- [#727](https://github.com/influxdata/telegraf/pull/727): riak input, thanks @jcoene!
- [#694](https://github.com/influxdata/telegraf/pull/694): DNS Query input, thanks @mjasion!
- [#724](https://github.com/influxdata/telegraf/pull/724): username matching for procstat input, thanks @zorel!
- [#736](https://github.com/influxdata/telegraf/pull/736): Ignore dummy filesystems from disk plugin. Thanks @PierreF!
- [#737](https://github.com/influxdata/telegraf/pull/737): Support multiple fields for statsd input. Thanks @mattheath!

### Bugfixes
- [#701](https://github.com/influxdata/telegraf/pull/701): output write count shouldnt print in quiet mode.
- [#746](https://github.com/influxdata/telegraf/pull/746): httpjson plugin: Fix HTTP GET parameters.

## v0.10.3 [2016-02-18]

### Release Notes
- Users of the `exec` and `kafka_consumer` (and the new `nats_consumer`
and `mqtt_consumer` plugins) can now specify the incoming data
format that they would like to parse. Currently supports: "json", "influx", and
"graphite"
- Users of message broker and file output plugins can now choose what data format
they would like to output. Currently supports: "influx" and "graphite"
- More info on parsing _incoming_ data formats can be found
[here](https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md)
- More info on serializing _outgoing_ data formats can be found
[here](https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md)
- Telegraf now has an option `flush_buffer_when_full` that will flush the
metric buffer whenever it fills up for each output, rather than dropping
points and only flushing on a set time interval. This will default to `true`
and is in the `[agent]` config section.

### Features
- [#652](https://github.com/influxdata/telegraf/pull/652): CouchDB Input Plugin. Thanks @codehate!
- [#655](https://github.com/influxdata/telegraf/pull/655): Support parsing arbitrary data formats. Currently limited to kafka_consumer and exec inputs.
- [#671](https://github.com/influxdata/telegraf/pull/671): Dovecot input plugin. Thanks @mikif70!
- [#680](https://github.com/influxdata/telegraf/pull/680): NATS consumer input plugin. Thanks @netixen!
- [#676](https://github.com/influxdata/telegraf/pull/676): MQTT consumer input plugin.
- [#683](https://github.com/influxdata/telegraf/pull/683): PostGRES input plugin: add pg_stat_bgwriter. Thanks @menardorama!
- [#679](https://github.com/influxdata/telegraf/pull/679): File/stdout output plugin.
- [#679](https://github.com/influxdata/telegraf/pull/679): Support for arbitrary output data formats.
- [#695](https://github.com/influxdata/telegraf/pull/695): raindrops input plugin. Thanks @burdandrei!
- [#650](https://github.com/influxdata/telegraf/pull/650): net_response input plugin. Thanks @titilambert!
- [#699](https://github.com/influxdata/telegraf/pull/699): Flush based on buffer size rather than time.
- [#682](https://github.com/influxdata/telegraf/pull/682): Mesos input plugin. Thanks @tripledes!

### Bugfixes
- [#443](https://github.com/influxdata/telegraf/issues/443): Fix Ping command timeout parameter on Linux.
- [#662](https://github.com/influxdata/telegraf/pull/667): Change `[tags]` to `[global_tags]` to fix multiple-plugin tags bug.
- [#642](https://github.com/influxdata/telegraf/issues/642): Riemann output plugin issues.
- [#394](https://github.com/influxdata/telegraf/issues/394): Support HTTP POST. Thanks @gabelev!
- [#715](https://github.com/influxdata/telegraf/pull/715): Fix influxdb precision config panic. Thanks @netixen!

## v0.10.2 [2016-02-04]

### Release Notes
- Statsd timing measurements are now aggregated into a single measurement with
fields.
- Graphite output now inserts tags into the bucket in alphabetical order.
- Normalized TLS/SSL support for output plugins: MQTT, AMQP, Kafka
- `verify_ssl` config option was removed from Kafka because it was actually
doing the opposite of what it claimed to do (yikes). It's been replaced by
`insecure_skip_verify`

### Features
- [#575](https://github.com/influxdata/telegraf/pull/575): Support for collecting Windows Performance Counters. Thanks @TheFlyingCorpse!
- [#564](https://github.com/influxdata/telegraf/issues/564): features for plugin writing simplification. Internal metric data type.
- [#603](https://github.com/influxdata/telegraf/pull/603): Aggregate statsd timing measurements into fields. Thanks @marcinbunsch!
- [#601](https://github.com/influxdata/telegraf/issues/601): Warn when overwriting cached metrics.
- [#614](https://github.com/influxdata/telegraf/pull/614): PowerDNS input plugin. Thanks @Kasen!
- [#617](https://github.com/influxdata/telegraf/pull/617): exec plugin: parse influx line protocol in addition to JSON.
- [#628](https://github.com/influxdata/telegraf/pull/628): Windows perf counters: pre-vista support

### Bugfixes
- [#595](https://github.com/influxdata/telegraf/issues/595): graphite output should include tags to separate duplicate measurements.
- [#599](https://github.com/influxdata/telegraf/issues/599): datadog plugin tags not working.
- [#600](https://github.com/influxdata/telegraf/issues/600): datadog measurement/field name parsing is wrong.
- [#602](https://github.com/influxdata/telegraf/issues/602): Fix statsd field name templating.
- [#612](https://github.com/influxdata/telegraf/pull/612): Docker input panic fix if stats received are nil.
- [#634](https://github.com/influxdata/telegraf/pull/634): Properly set host headers in httpjson. Thanks @reginaldosousa!

## v0.10.1 [2016-01-27]

### Release Notes

- Telegraf now keeps a fixed-length buffer of metrics per-output. This buffer
defaults to 10,000 metrics, and is adjustable. The buffer is cleared when a
successful write to that output occurs.
- The docker plugin has been significantly overhauled to add more metrics
and allow for docker-machine (incl OSX) support.
[See the readme](https://github.com/influxdata/telegraf/blob/master/plugins/inputs/docker/README.md)
for the latest measurements, fields, and tags. There is also now support for
specifying a docker endpoint to get metrics from.

### Features
- [#509](https://github.com/influxdata/telegraf/pull/509): Flatten JSON arrays with indices. Thanks @psilva261!
- [#512](https://github.com/influxdata/telegraf/pull/512): Python 3 build script, add lsof dep to package. Thanks @Ormod!
- [#475](https://github.com/influxdata/telegraf/pull/475): Add response time to httpjson plugin. Thanks @titilambert!
- [#519](https://github.com/influxdata/telegraf/pull/519): Added a sensors input based on lm-sensors. Thanks @md14454!
- [#467](https://github.com/influxdata/telegraf/issues/467): Add option to disable statsd measurement name conversion.
- [#534](https://github.com/influxdata/telegraf/pull/534): NSQ input plugin. Thanks @allingeek!
- [#494](https://github.com/influxdata/telegraf/pull/494): Graphite output plugin. Thanks @titilambert!
- AMQP SSL support. Thanks @ekini!
- [#539](https://github.com/influxdata/telegraf/pull/539): Reload config on SIGHUP. Thanks @titilambert!
- [#522](https://github.com/influxdata/telegraf/pull/522): Phusion passenger input plugin. Thanks @kureikain!
- [#541](https://github.com/influxdata/telegraf/pull/541): Kafka output TLS cert support. Thanks @Ormod!
- [#551](https://github.com/influxdata/telegraf/pull/551): Statsd UDP read packet size now defaults to 1500 bytes, and is configurable.
- [#552](https://github.com/influxdata/telegraf/pull/552): Support for collection interval jittering.
- [#484](https://github.com/influxdata/telegraf/issues/484): Include usage percent with procstat metrics.
- [#553](https://github.com/influxdata/telegraf/pull/553): Amazon CloudWatch output. thanks @skwong2!
- [#503](https://github.com/influxdata/telegraf/pull/503): Support docker endpoint configuration.
- [#563](https://github.com/influxdata/telegraf/pull/563): Docker plugin overhaul.
- [#285](https://github.com/influxdata/telegraf/issues/285): Fixed-size buffer of points.
- [#546](https://github.com/influxdata/telegraf/pull/546): SNMP Input plugin. Thanks @titilambert!
- [#589](https://github.com/influxdata/telegraf/pull/589): Microsoft SQL Server input plugin. Thanks @zensqlmonitor!
- [#573](https://github.com/influxdata/telegraf/pull/573): Github webhooks consumer input. Thanks @jackzampolin!
- [#471](https://github.com/influxdata/telegraf/pull/471): httpjson request headers. Thanks @asosso!

### Bugfixes
- [#506](https://github.com/influxdata/telegraf/pull/506): Ping input doesn't return response time metric when timeout. Thanks @titilambert!
- [#508](https://github.com/influxdata/telegraf/pull/508): Fix prometheus cardinality issue with the `net` plugin
- [#499](https://github.com/influxdata/telegraf/issues/499) & [#502](https://github.com/influxdata/telegraf/issues/502): php fpm unix socket and other fixes, thanks @kureikain!
- [#543](https://github.com/influxdata/telegraf/issues/543): Statsd Packet size sometimes truncated.
- [#440](https://github.com/influxdata/telegraf/issues/440): Don't query filtered devices for disk stats.
- [#463](https://github.com/influxdata/telegraf/issues/463): Docker plugin not working on AWS Linux
- [#568](https://github.com/influxdata/telegraf/issues/568): Multiple output race condition.
- [#585](https://github.com/influxdata/telegraf/pull/585): Log stack trace and continue on Telegraf panic. Thanks @wutaizeng!

## v0.10.0 [2016-01-12]

### Release Notes
- Linux packages have been taken out of `opt`, the binary is now in `/usr/bin`
and configuration files are in `/etc/telegraf`
- **breaking change** `plugins` have been renamed to `inputs`. This was done because
`plugins` is too generic, as there are now also "output plugins", and will likely
be "aggregator plugins" and "filter plugins" in the future. Additionally,
`inputs/` and `outputs/` directories have been placed in the root-level `plugins/`
directory.
- **breaking change** the `io` plugin has been renamed `diskio`
- **breaking change** plugin measurements aggregated into a single measurement.
- **breaking change** `jolokia` plugin: must use global tag/drop/pass parameters
for configuration.
- **breaking change** `twemproxy` plugin: `prefix` option removed.
- **breaking change** `procstat` cpu measurements are now prepended with `cpu_time_`
instead of only `cpu_`
- **breaking change** some command-line flags have been renamed to separate words.
`-configdirectory` -> `-config-directory`, `-filter` -> `-input-filter`,
`-outputfilter` -> `-output-filter`
- The prometheus plugin schema has not been changed (measurements have not been
aggregated).

### Packaging change note:

RHEL/CentOS users upgrading from 0.2.x to 0.10.0 will probably have their
configurations overwritten by the upgrade. There is a backup stored at
/etc/telegraf/telegraf.conf.$(date +%s).backup.

### Features
- Plugin measurements aggregated into a single measurement.
- Added ability to specify per-plugin tags
- Added ability to specify per-plugin measurement suffix and prefix.
(`name_prefix` and `name_suffix`)
- Added ability to override base plugin measurement name. (`name_override`)

### Bugfixes

## v0.2.5 [unreleased]

### Features
- [#427](https://github.com/influxdata/telegraf/pull/427): zfs plugin: pool stats added. Thanks @allenpetersen!
- [#428](https://github.com/influxdata/telegraf/pull/428): Amazon Kinesis output. Thanks @jimmystewpot!
- [#449](https://github.com/influxdata/telegraf/pull/449): influxdb plugin, thanks @mark-rushakoff

### Bugfixes
- [#430](https://github.com/influxdata/telegraf/issues/430): Network statistics removed in elasticsearch 2.1. Thanks @jipperinbham!
- [#452](https://github.com/influxdata/telegraf/issues/452): Elasticsearch open file handles error. Thanks @jipperinbham!

## v0.2.4 [2015-12-08]

### Features
- [#412](https://github.com/influxdata/telegraf/pull/412): Additional memcached stats. Thanks @mgresser!
- [#410](https://github.com/influxdata/telegraf/pull/410): Additional redis metrics. Thanks @vlaadbrain!
- [#414](https://github.com/influxdata/telegraf/issues/414): Jolokia plugin auth parameters
- [#415](https://github.com/influxdata/telegraf/issues/415): memcached plugin: support unix sockets
- [#418](https://github.com/influxdata/telegraf/pull/418): memcached plugin additional unit tests.
- [#408](https://github.com/influxdata/telegraf/pull/408): MailChimp plugin.
- [#382](https://github.com/influxdata/telegraf/pull/382): Add system wide network protocol stats to `net` plugin.
- [#401](https://github.com/influxdata/telegraf/pull/401): Support pass/drop/tagpass/tagdrop for outputs. Thanks @oldmantaiter!

### Bugfixes
- [#405](https://github.com/influxdata/telegraf/issues/405): Prometheus output cardinality issue
- [#388](https://github.com/influxdata/telegraf/issues/388): Fix collection hangup when cpu times decrement.

## v0.2.3 [2015-11-30]

### Release Notes
- **breaking change** The `kafka` plugin has been renamed to `kafka_consumer`.
and most of the config option names have changed.
This only affects the kafka consumer _plugin_ (not the
output). There were a number of problems with the kafka plugin that led to it
only collecting data once at startup, so the kafka plugin was basically non-
functional.
- Plugins can now be specified as a list, and multiple plugin instances of the
same type can be specified, like this:

```
[[inputs.cpu]]
  percpu = false
  totalcpu = true

[[inputs.cpu]]
  percpu = true
  totalcpu = false
  drop = ["cpu_time"]
```

- Riemann output added
- Aerospike plugin: tag changed from `host` -> `aerospike_host`

### Features
- [#379](https://github.com/influxdata/telegraf/pull/379): Riemann output, thanks @allenj!
- [#375](https://github.com/influxdata/telegraf/pull/375): kafka_consumer service plugin.
- [#392](https://github.com/influxdata/telegraf/pull/392): Procstat plugin can now accept pgrep -f pattern, thanks @ecarreras!
- [#383](https://github.com/influxdata/telegraf/pull/383): Specify plugins as a list.
- [#354](https://github.com/influxdata/telegraf/pull/354): Add ability to specify multiple metrics in one statsd line. Thanks @MerlinDMC!

### Bugfixes
- [#371](https://github.com/influxdata/telegraf/issues/371): Kafka consumer plugin not functioning.
- [#389](https://github.com/influxdata/telegraf/issues/389): NaN value panic

## v0.2.2 [2015-11-18]

### Release Notes
- 0.2.1 has a bug where all lists within plugins get duplicated, this includes
lists of servers/URLs. 0.2.2 is being released solely to fix that bug

### Bugfixes
- [#377](https://github.com/influxdata/telegraf/pull/377): Fix for duplicate slices in inputs.

## v0.2.1 [2015-11-16]

### Release Notes
- Telegraf will no longer use docker-compose for "long" unit test, it has been
changed to just run docker commands in the Makefile. See `make docker-run` and
`make docker-kill`. `make test` will still run all unit tests with docker.
- Long unit tests are now run in CircleCI, with docker & race detector
- Redis plugin tag has changed from `host` to `server`
- HAProxy plugin tag has changed from `host` to `server`
- UDP output now supported
- Telegraf will now compile on FreeBSD
- Users can now specify outputs as lists, specifying multiple outputs of the
same type.

### Features
- [#325](https://github.com/influxdata/telegraf/pull/325): NSQ output. Thanks @jrxFive!
- [#318](https://github.com/influxdata/telegraf/pull/318): Prometheus output. Thanks @oldmantaiter!
- [#338](https://github.com/influxdata/telegraf/pull/338): Restart Telegraf on package upgrade. Thanks @linsomniac!
- [#337](https://github.com/influxdata/telegraf/pull/337): Jolokia plugin, thanks @saiello!
- [#350](https://github.com/influxdata/telegraf/pull/350): Amon output.
- [#365](https://github.com/influxdata/telegraf/pull/365): Twemproxy plugin by @codeb2cc
- [#317](https://github.com/influxdata/telegraf/issues/317): ZFS plugin, thanks @cornerot!
- [#364](https://github.com/influxdata/telegraf/pull/364): Support InfluxDB UDP output.
- [#370](https://github.com/influxdata/telegraf/pull/370): Support specifying multiple outputs, as lists.
- [#372](https://github.com/influxdata/telegraf/pull/372): Remove gosigar and update go-dockerclient for FreeBSD support. Thanks @MerlinDMC!

### Bugfixes
- [#331](https://github.com/influxdata/telegraf/pull/331): Dont overwrite host tag in redis plugin.
- [#336](https://github.com/influxdata/telegraf/pull/336): Mongodb plugin should take 2 measurements.
- [#351](https://github.com/influxdata/telegraf/issues/317): Fix continual "CREATE DATABASE" in writes
- [#360](https://github.com/influxdata/telegraf/pull/360): Apply prefix before ShouldPass check. Thanks @sotfo!

## v0.2.0 [2015-10-27]

### Release Notes
- The -test flag will now only output 2 collections for plugins that need it
- There is a new agent configuration option: `flush_interval`. This option tells
Telegraf how often to flush data to InfluxDB and other output sinks. For example,
users can set `interval = "2s"` and `flush_interval = "60s"` for Telegraf to
collect data every 2 seconds, and flush every 60 seconds.
- `precision` and `utc` are no longer valid agent config values. `precision` has
moved to the `influxdb` output config, where it will continue to default to "s"
- debug and test output will now print the raw line-protocol string
- Telegraf will now, by default, round the collection interval to the nearest
even interval. This means that `interval="10s"` will collect every :00, :10, etc.
To ease scale concerns, flushing will be "jittered" by a random amount so that
all Telegraf instances do not flush at the same time. Both of these options can
be controlled via the `round_interval` and `flush_jitter` config options.
- Telegraf will now retry metric flushes twice

### Features
- [#205](https://github.com/influxdata/telegraf/issues/205): Include per-db redis keyspace info
- [#226](https://github.com/influxdata/telegraf/pull/226): Add timestamps to points in Kafka/AMQP outputs. Thanks @ekini
- [#90](https://github.com/influxdata/telegraf/issues/90): Add Docker labels to tags in docker plugin
- [#223](https://github.com/influxdata/telegraf/pull/223): Add port tag to nginx plugin. Thanks @neezgee!
- [#227](https://github.com/influxdata/telegraf/pull/227): Add command intervals to exec plugin. Thanks @jpalay!
- [#241](https://github.com/influxdata/telegraf/pull/241): MQTT Output. Thanks @shirou!
- Memory plugin: cached and buffered measurements re-added
- Logging: additional logging for each collection interval, track the number
of metrics collected and from how many inputs.
- [#240](https://github.com/influxdata/telegraf/pull/240): procstat plugin, thanks @ranjib!
- [#244](https://github.com/influxdata/telegraf/pull/244): netstat plugin, thanks @shirou!
- [#262](https://github.com/influxdata/telegraf/pull/262): zookeeper plugin, thanks @jrxFive!
- [#237](https://github.com/influxdata/telegraf/pull/237): statsd service plugin, thanks @sparrc
- [#273](https://github.com/influxdata/telegraf/pull/273): puppet agent plugin, thats @jrxFive!
- [#280](https://github.com/influxdata/telegraf/issues/280): Use InfluxDB client v2.
- [#281](https://github.com/influxdata/telegraf/issues/281): Eliminate need to deep copy Batch Points.
- [#286](https://github.com/influxdata/telegraf/issues/286): bcache plugin, thanks @cornerot!
- [#287](https://github.com/influxdata/telegraf/issues/287): Batch AMQP output, thanks @ekini!
- [#301](https://github.com/influxdata/telegraf/issues/301): Collect on even intervals
- [#298](https://github.com/influxdata/telegraf/pull/298): Support retrying output writes
- [#300](https://github.com/influxdata/telegraf/issues/300): aerospike plugin. Thanks @oldmantaiter!
- [#322](https://github.com/influxdata/telegraf/issues/322): Librato output. Thanks @jipperinbham!

### Bugfixes
- [#228](https://github.com/influxdata/telegraf/pull/228): New version of package will replace old one. Thanks @ekini!
- [#232](https://github.com/influxdata/telegraf/pull/232): Fix bashism run during deb package installation. Thanks @yankcrime!
- [#261](https://github.com/influxdata/telegraf/issues/260): RabbitMQ panics if wrong credentials given. Thanks @ekini!
- [#245](https://github.com/influxdata/telegraf/issues/245): Document Exec plugin example. Thanks @ekini!
- [#264](https://github.com/influxdata/telegraf/issues/264): logrotate config file fixes. Thanks @linsomniac!
- [#290](https://github.com/influxdata/telegraf/issues/290): Fix some plugins sending their values as strings.
- [#289](https://github.com/influxdata/telegraf/issues/289): Fix accumulator panic on nil tags.
- [#302](https://github.com/influxdata/telegraf/issues/302): Fix `[tags]` getting applied, thanks @gotyaoi!

## v0.1.9 [2015-09-22]

### Release Notes
- InfluxDB output config change: `url` is now `urls`, and is a list. Config files
will still be backwards compatible if only `url` is specified.
- The -test flag will now output two metric collections
- Support for filtering telegraf outputs on the CLI -- Telegraf will now
allow filtering of output sinks on the command-line using the `-outputfilter`
flag, much like how the `-filter` flag works for inputs.
- Support for filtering on config-file creation -- Telegraf now supports
filtering to -sample-config command. You can now run
`telegraf -sample-config -filter cpu -outputfilter influxdb` to get a config
file with only the cpu plugin defined, and the influxdb output defined.
- **Breaking Change**: The CPU collection plugin has been refactored to fix some
bugs and outdated dependency issues. At the same time, I also decided to fix
a naming consistency issue, so cpu_percentageIdle will become cpu_usage_idle.
Also, all CPU time measurements now have it indicated in their name, so cpu_idle
will become cpu_time_idle. Additionally, cpu_time measurements are going to be
dropped in the default config.
- **Breaking Change**: The memory plugin has been refactored and some measurements
have been renamed for consistency. Some measurements have also been removed from being outputted. They are still being collected by gopsutil, and could easily be
re-added in a "verbose" mode if there is demand for it.

### Features
- [#143](https://github.com/influxdata/telegraf/issues/143): InfluxDB clustering support
- [#181](https://github.com/influxdata/telegraf/issues/181): Makefile GOBIN support. Thanks @Vye!
- [#203](https://github.com/influxdata/telegraf/pull/200): AMQP output. Thanks @ekini!
- [#182](https://github.com/influxdata/telegraf/pull/182): OpenTSDB output. Thanks @rplessl!
- [#187](https://github.com/influxdata/telegraf/pull/187): Retry output sink connections on startup.
- [#220](https://github.com/influxdata/telegraf/pull/220): Add port tag to apache plugin. Thanks @neezgee!
- [#217](https://github.com/influxdata/telegraf/pull/217): Add filtering for output sinks
and filtering when specifying a config file.

### Bugfixes
- [#170](https://github.com/influxdata/telegraf/issues/170): Systemd support
- [#175](https://github.com/influxdata/telegraf/issues/175): Set write precision before gathering metrics
- [#178](https://github.com/influxdata/telegraf/issues/178): redis plugin, multiple server thread hang bug
- Fix net plugin on darwin
- [#84](https://github.com/influxdata/telegraf/issues/84): Fix docker plugin on CentOS. Thanks @neezgee!
- [#189](https://github.com/influxdata/telegraf/pull/189): Fix mem_used_perc. Thanks @mced!
- [#192](https://github.com/influxdata/telegraf/issues/192): Increase compatibility of postgresql plugin. Now supports versions 8.1+
- [#203](https://github.com/influxdata/telegraf/issues/203): EL5 rpm support. Thanks @ekini!
- [#206](https://github.com/influxdata/telegraf/issues/206): CPU steal/guest values wrong on linux.
- [#212](https://github.com/influxdata/telegraf/issues/212): Add hashbang to postinstall script. Thanks @ekini!
- [#212](https://github.com/influxdata/telegraf/issues/212): Fix makefile warning. Thanks @ekini!

## v0.1.8 [2015-09-04]

### Release Notes
- Telegraf will now write data in UTC at second precision by default
- Now using Go 1.5 to build telegraf

### Features
- [#150](https://github.com/influxdata/telegraf/pull/150): Add Host Uptime metric to system plugin
- [#158](https://github.com/influxdata/telegraf/pull/158): Apache Plugin. Thanks @KPACHbIuLLIAnO4
- [#159](https://github.com/influxdata/telegraf/pull/159): Use second precision for InfluxDB writes
- [#165](https://github.com/influxdata/telegraf/pull/165): Add additional metrics to mysql plugin. Thanks @nickscript0
- [#162](https://github.com/influxdata/telegraf/pull/162): Write UTC by default, provide option
- [#166](https://github.com/influxdata/telegraf/pull/166): Upload binaries to S3
- [#169](https://github.com/influxdata/telegraf/pull/169): Ping plugin

### Bugfixes

## v0.1.7 [2015-08-28]

### Features
- [#38](https://github.com/influxdata/telegraf/pull/38): Kafka output producer.
- [#133](https://github.com/influxdata/telegraf/pull/133): Add plugin.Gather error logging. Thanks @nickscript0!
- [#136](https://github.com/influxdata/telegraf/issues/136): Add a -usage flag for printing usage of a single plugin.
- [#137](https://github.com/influxdata/telegraf/issues/137): Memcached: fix when a value contains a space
- [#138](https://github.com/influxdata/telegraf/issues/138): MySQL server address tag.
- [#142](https://github.com/influxdata/telegraf/pull/142): Add Description and SampleConfig funcs to output interface
- Indent the toml config file for readability

### Bugfixes
- [#128](https://github.com/influxdata/telegraf/issues/128): system_load measurement missing.
- [#129](https://github.com/influxdata/telegraf/issues/129): Latest pkg url fix.
- [#131](https://github.com/influxdata/telegraf/issues/131): Fix memory reporting on linux & darwin. Thanks @subhachandrachandra!
- [#140](https://github.com/influxdata/telegraf/issues/140): Memory plugin prec->perc typo fix. Thanks @brunoqc!

## v0.1.6 [2015-08-20]

### Features
- [#112](https://github.com/influxdata/telegraf/pull/112): Datadog output. Thanks @jipperinbham!
- [#116](https://github.com/influxdata/telegraf/pull/116): Use godep to vendor all dependencies
- [#120](https://github.com/influxdata/telegraf/pull/120): Httpjson plugin. Thanks @jpalay & @alvaromorales!

### Bugfixes
- [#113](https://github.com/influxdata/telegraf/issues/113): Update README with Telegraf/InfluxDB compatibility
- [#118](https://github.com/influxdata/telegraf/pull/118): Fix for disk usage stats in Windows. Thanks @srfraser!
- [#122](https://github.com/influxdata/telegraf/issues/122): Fix for DiskUsage segv fault. Thanks @srfraser!
- [#126](https://github.com/influxdata/telegraf/issues/126): Nginx plugin not catching net.SplitHostPort error

## v0.1.5 [2015-08-13]

### Features
- [#54](https://github.com/influxdata/telegraf/pull/54): MongoDB plugin. Thanks @jipperinbham!
- [#55](https://github.com/influxdata/telegraf/pull/55): Elasticsearch plugin. Thanks @brocaar!
- [#71](https://github.com/influxdata/telegraf/pull/71): HAProxy plugin. Thanks @kureikain!
- [#72](https://github.com/influxdata/telegraf/pull/72): Adding TokuDB metrics to MySQL. Thanks vadimtk!
- [#73](https://github.com/influxdata/telegraf/pull/73): RabbitMQ plugin. Thanks @ianunruh!
- [#77](https://github.com/influxdata/telegraf/issues/77): Automatically create database.
- [#79](https://github.com/influxdata/telegraf/pull/56): Nginx plugin. Thanks @codeb2cc!
- [#86](https://github.com/influxdata/telegraf/pull/86): Lustre2 plugin. Thanks srfraser!
- [#91](https://github.com/influxdata/telegraf/pull/91): Unit testing
- [#92](https://github.com/influxdata/telegraf/pull/92): Exec plugin. Thanks @alvaromorales!
- [#98](https://github.com/influxdata/telegraf/pull/98): LeoFS plugin. Thanks @mocchira!
- [#103](https://github.com/influxdata/telegraf/pull/103): Filter by metric tags. Thanks @srfraser!
- [#106](https://github.com/influxdata/telegraf/pull/106): Options to filter plugins on startup. Thanks @zepouet!
- [#107](https://github.com/influxdata/telegraf/pull/107): Multiple outputs beyong influxdb. Thanks @jipperinbham!
- [#108](https://github.com/influxdata/telegraf/issues/108): Support setting per-CPU and total-CPU gathering.
- [#111](https://github.com/influxdata/telegraf/pull/111): Report CPU Usage in cpu plugin. Thanks @jpalay!

### Bugfixes
- [#85](https://github.com/influxdata/telegraf/pull/85): Fix GetLocalHost testutil function for mac users
- [#89](https://github.com/influxdata/telegraf/pull/89): go fmt fixes
- [#94](https://github.com/influxdata/telegraf/pull/94): Fix for issue #93, explicitly call sarama.v1 -> sarama
- [#101](https://github.com/influxdata/telegraf/issues/101): switch back from master branch if building locally
- [#99](https://github.com/influxdata/telegraf/issues/99): update integer output to new InfluxDB line protocol format

## v0.1.4 [2015-07-09]

### Features
- [#56](https://github.com/influxdata/telegraf/pull/56): Update README for Kafka plugin. Thanks @EmilS!

### Bugfixes
- [#50](https://github.com/influxdata/telegraf/pull/50): Fix init.sh script to use telegraf directory. Thanks @jseriff!
- [#52](https://github.com/influxdata/telegraf/pull/52): Update CHANGELOG to reference updated directory. Thanks @benfb!

## v0.1.3 [2015-07-05]

### Features
- [#35](https://github.com/influxdata/telegraf/pull/35): Add Kafka plugin. Thanks @EmilS!
- [#47](https://github.com/influxdata/telegraf/pull/47): Add RethinkDB plugin. Thanks @jipperinbham!

### Bugfixes
- [#45](https://github.com/influxdata/telegraf/pull/45): Skip disk tags that don't have a value. Thanks @jhofeditz!
- [#43](https://github.com/influxdata/telegraf/pull/43): Fix bug in MySQL plugin. Thanks @marcosnils!

## v0.1.2 [2015-07-01]

### Features
- [#12](https://github.com/influxdata/telegraf/pull/12): Add Linux/ARM to the list of built binaries. Thanks @voxxit!
- [#14](https://github.com/influxdata/telegraf/pull/14): Clarify the S3 buckets that Telegraf is pushed to.
- [#16](https://github.com/influxdata/telegraf/pull/16): Convert Redis to use URI, support Redis AUTH. Thanks @jipperinbham!
- [#21](https://github.com/influxdata/telegraf/pull/21): Add memcached plugin. Thanks @Yukki!

### Bugfixes
- [#13](https://github.com/influxdata/telegraf/pull/13): Fix the packaging script.
- [#19](https://github.com/influxdata/telegraf/pull/19): Add host name to metric tags. Thanks @sherifzain!
- [#20](https://github.com/influxdata/telegraf/pull/20): Fix race condition with accumulator mutex. Thanks @nkatsaros!
- [#23](https://github.com/influxdata/telegraf/pull/23): Change name of folder for packages. Thanks @colinrymer!
- [#32](https://github.com/influxdata/telegraf/pull/32): Fix spelling of memoory -> memory. Thanks @tylernisonoff!

## v0.1.1 [2015-06-19]

### Release Notes

This is the initial release of Telegraf.
