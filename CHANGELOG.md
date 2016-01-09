## v0.10.0 [unreleased]

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
- [#427](https://github.com/influxdb/telegraf/pull/427): zfs plugin: pool stats added. Thanks @allenpetersen!
- [#428](https://github.com/influxdb/telegraf/pull/428): Amazon Kinesis output. Thanks @jimmystewpot!
- [#449](https://github.com/influxdb/telegraf/pull/449): influxdb plugin, thanks @mark-rushakoff

### Bugfixes
- [#430](https://github.com/influxdb/telegraf/issues/430): Network statistics removed in elasticsearch 2.1. Thanks @jipperinbham!
- [#452](https://github.com/influxdb/telegraf/issues/452): Elasticsearch open file handles error. Thanks @jipperinbham!

## v0.2.4 [2015-12-08]

### Features
- [#412](https://github.com/influxdb/telegraf/pull/412): Additional memcached stats. Thanks @mgresser!
- [#410](https://github.com/influxdb/telegraf/pull/410): Additional redis metrics. Thanks @vlaadbrain!
- [#414](https://github.com/influxdb/telegraf/issues/414): Jolokia plugin auth parameters
- [#415](https://github.com/influxdb/telegraf/issues/415): memcached plugin: support unix sockets
- [#418](https://github.com/influxdb/telegraf/pull/418): memcached plugin additional unit tests.
- [#408](https://github.com/influxdb/telegraf/pull/408): MailChimp plugin.
- [#382](https://github.com/influxdb/telegraf/pull/382): Add system wide network protocol stats to `net` plugin.
- [#401](https://github.com/influxdb/telegraf/pull/401): Support pass/drop/tagpass/tagdrop for outputs. Thanks @oldmantaiter!

### Bugfixes
- [#405](https://github.com/influxdb/telegraf/issues/405): Prometheus output cardinality issue
- [#388](https://github.com/influxdb/telegraf/issues/388): Fix collection hangup when cpu times decrement.

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
- [#379](https://github.com/influxdb/telegraf/pull/379): Riemann output, thanks @allenj!
- [#375](https://github.com/influxdb/telegraf/pull/375): kafka_consumer service plugin.
- [#392](https://github.com/influxdb/telegraf/pull/392): Procstat plugin can now accept pgrep -f pattern, thanks @ecarreras!
- [#383](https://github.com/influxdb/telegraf/pull/383): Specify plugins as a list.
- [#354](https://github.com/influxdb/telegraf/pull/354): Add ability to specify multiple metrics in one statsd line. Thanks @MerlinDMC!

### Bugfixes
- [#371](https://github.com/influxdb/telegraf/issues/371): Kafka consumer plugin not functioning.
- [#389](https://github.com/influxdb/telegraf/issues/389): NaN value panic

## v0.2.2 [2015-11-18]

### Release Notes
- 0.2.1 has a bug where all lists within plugins get duplicated, this includes
lists of servers/URLs. 0.2.2 is being released solely to fix that bug

### Bugfixes
- [#377](https://github.com/influxdb/telegraf/pull/377): Fix for duplicate slices in inputs.

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
- [#325](https://github.com/influxdb/telegraf/pull/325): NSQ output. Thanks @jrxFive!
- [#318](https://github.com/influxdb/telegraf/pull/318): Prometheus output. Thanks @oldmantaiter!
- [#338](https://github.com/influxdb/telegraf/pull/338): Restart Telegraf on package upgrade. Thanks @linsomniac!
- [#337](https://github.com/influxdb/telegraf/pull/337): Jolokia plugin, thanks @saiello!
- [#350](https://github.com/influxdb/telegraf/pull/350): Amon output.
- [#365](https://github.com/influxdb/telegraf/pull/365): Twemproxy plugin by @codeb2cc
- [#317](https://github.com/influxdb/telegraf/issues/317): ZFS plugin, thanks @cornerot!
- [#364](https://github.com/influxdb/telegraf/pull/364): Support InfluxDB UDP output.
- [#370](https://github.com/influxdb/telegraf/pull/370): Support specifying multiple outputs, as lists.
- [#372](https://github.com/influxdb/telegraf/pull/372): Remove gosigar and update go-dockerclient for FreeBSD support. Thanks @MerlinDMC!

### Bugfixes
- [#331](https://github.com/influxdb/telegraf/pull/331): Dont overwrite host tag in redis plugin.
- [#336](https://github.com/influxdb/telegraf/pull/336): Mongodb plugin should take 2 measurements.
- [#351](https://github.com/influxdb/telegraf/issues/317): Fix continual "CREATE DATABASE" in writes
- [#360](https://github.com/influxdb/telegraf/pull/360): Apply prefix before ShouldPass check. Thanks @sotfo!

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
- [#205](https://github.com/influxdb/telegraf/issues/205): Include per-db redis keyspace info
- [#226](https://github.com/influxdb/telegraf/pull/226): Add timestamps to points in Kafka/AMQP outputs. Thanks @ekini
- [#90](https://github.com/influxdb/telegraf/issues/90): Add Docker labels to tags in docker plugin
- [#223](https://github.com/influxdb/telegraf/pull/223): Add port tag to nginx plugin. Thanks @neezgee!
- [#227](https://github.com/influxdb/telegraf/pull/227): Add command intervals to exec plugin. Thanks @jpalay!
- [#241](https://github.com/influxdb/telegraf/pull/241): MQTT Output. Thanks @shirou!
- Memory plugin: cached and buffered measurements re-added
- Logging: additional logging for each collection interval, track the number
of metrics collected and from how many inputs.
- [#240](https://github.com/influxdb/telegraf/pull/240): procstat plugin, thanks @ranjib!
- [#244](https://github.com/influxdb/telegraf/pull/244): netstat plugin, thanks @shirou!
- [#262](https://github.com/influxdb/telegraf/pull/262): zookeeper plugin, thanks @jrxFive!
- [#237](https://github.com/influxdb/telegraf/pull/237): statsd service plugin, thanks @sparrc
- [#273](https://github.com/influxdb/telegraf/pull/273): puppet agent plugin, thats @jrxFive!
- [#280](https://github.com/influxdb/telegraf/issues/280): Use InfluxDB client v2.
- [#281](https://github.com/influxdb/telegraf/issues/281): Eliminate need to deep copy Batch Points.
- [#286](https://github.com/influxdb/telegraf/issues/286): bcache plugin, thanks @cornerot!
- [#287](https://github.com/influxdb/telegraf/issues/287): Batch AMQP output, thanks @ekini!
- [#301](https://github.com/influxdb/telegraf/issues/301): Collect on even intervals
- [#298](https://github.com/influxdb/telegraf/pull/298): Support retrying output writes
- [#300](https://github.com/influxdb/telegraf/issues/300): aerospike plugin. Thanks @oldmantaiter!
- [#322](https://github.com/influxdb/telegraf/issues/322): Librato output. Thanks @jipperinbham!

### Bugfixes
- [#228](https://github.com/influxdb/telegraf/pull/228): New version of package will replace old one. Thanks @ekini!
- [#232](https://github.com/influxdb/telegraf/pull/232): Fix bashism run during deb package installation. Thanks @yankcrime!
- [#261](https://github.com/influxdb/telegraf/issues/260): RabbitMQ panics if wrong credentials given. Thanks @ekini!
- [#245](https://github.com/influxdb/telegraf/issues/245): Document Exec plugin example. Thanks @ekini!
- [#264](https://github.com/influxdb/telegraf/issues/264): logrotate config file fixes. Thanks @linsomniac!
- [#290](https://github.com/influxdb/telegraf/issues/290): Fix some plugins sending their values as strings.
- [#289](https://github.com/influxdb/telegraf/issues/289): Fix accumulator panic on nil tags.
- [#302](https://github.com/influxdb/telegraf/issues/302): Fix `[tags]` getting applied, thanks @gotyaoi!

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
- [#143](https://github.com/influxdb/telegraf/issues/143): InfluxDB clustering support
- [#181](https://github.com/influxdb/telegraf/issues/181): Makefile GOBIN support. Thanks @Vye!
- [#203](https://github.com/influxdb/telegraf/pull/200): AMQP output. Thanks @ekini!
- [#182](https://github.com/influxdb/telegraf/pull/182): OpenTSDB output. Thanks @rplessl!
- [#187](https://github.com/influxdb/telegraf/pull/187): Retry output sink connections on startup.
- [#220](https://github.com/influxdb/telegraf/pull/220): Add port tag to apache plugin. Thanks @neezgee!
- [#217](https://github.com/influxdb/telegraf/pull/217): Add filtering for output sinks
and filtering when specifying a config file.

### Bugfixes
- [#170](https://github.com/influxdb/telegraf/issues/170): Systemd support
- [#175](https://github.com/influxdb/telegraf/issues/175): Set write precision before gathering metrics
- [#178](https://github.com/influxdb/telegraf/issues/178): redis plugin, multiple server thread hang bug
- Fix net plugin on darwin
- [#84](https://github.com/influxdb/telegraf/issues/84): Fix docker plugin on CentOS. Thanks @neezgee!
- [#189](https://github.com/influxdb/telegraf/pull/189): Fix mem_used_perc. Thanks @mced!
- [#192](https://github.com/influxdb/telegraf/issues/192): Increase compatibility of postgresql plugin. Now supports versions 8.1+
- [#203](https://github.com/influxdb/telegraf/issues/203): EL5 rpm support. Thanks @ekini!
- [#206](https://github.com/influxdb/telegraf/issues/206): CPU steal/guest values wrong on linux.
- [#212](https://github.com/influxdb/telegraf/issues/212): Add hashbang to postinstall script. Thanks @ekini!
- [#212](https://github.com/influxdb/telegraf/issues/212): Fix makefile warning. Thanks @ekini!

## v0.1.8 [2015-09-04]

### Release Notes
- Telegraf will now write data in UTC at second precision by default
- Now using Go 1.5 to build telegraf

### Features
- [#150](https://github.com/influxdb/telegraf/pull/150): Add Host Uptime metric to system plugin
- [#158](https://github.com/influxdb/telegraf/pull/158): Apache Plugin. Thanks @KPACHbIuLLIAnO4
- [#159](https://github.com/influxdb/telegraf/pull/159): Use second precision for InfluxDB writes
- [#165](https://github.com/influxdb/telegraf/pull/165): Add additional metrics to mysql plugin. Thanks @nickscript0
- [#162](https://github.com/influxdb/telegraf/pull/162): Write UTC by default, provide option
- [#166](https://github.com/influxdb/telegraf/pull/166): Upload binaries to S3
- [#169](https://github.com/influxdb/telegraf/pull/169): Ping plugin

### Bugfixes

## v0.1.7 [2015-08-28]

### Features
- [#38](https://github.com/influxdb/telegraf/pull/38): Kafka output producer.
- [#133](https://github.com/influxdb/telegraf/pull/133): Add plugin.Gather error logging. Thanks @nickscript0!
- [#136](https://github.com/influxdb/telegraf/issues/136): Add a -usage flag for printing usage of a single plugin.
- [#137](https://github.com/influxdb/telegraf/issues/137): Memcached: fix when a value contains a space
- [#138](https://github.com/influxdb/telegraf/issues/138): MySQL server address tag.
- [#142](https://github.com/influxdb/telegraf/pull/142): Add Description and SampleConfig funcs to output interface
- Indent the toml config file for readability

### Bugfixes
- [#128](https://github.com/influxdb/telegraf/issues/128): system_load measurement missing.
- [#129](https://github.com/influxdb/telegraf/issues/129): Latest pkg url fix.
- [#131](https://github.com/influxdb/telegraf/issues/131): Fix memory reporting on linux & darwin. Thanks @subhachandrachandra!
- [#140](https://github.com/influxdb/telegraf/issues/140): Memory plugin prec->perc typo fix. Thanks @brunoqc!

## v0.1.6 [2015-08-20]

### Features
- [#112](https://github.com/influxdb/telegraf/pull/112): Datadog output. Thanks @jipperinbham!
- [#116](https://github.com/influxdb/telegraf/pull/116): Use godep to vendor all dependencies
- [#120](https://github.com/influxdb/telegraf/pull/120): Httpjson plugin. Thanks @jpalay & @alvaromorales!

### Bugfixes
- [#113](https://github.com/influxdb/telegraf/issues/113): Update README with Telegraf/InfluxDB compatibility
- [#118](https://github.com/influxdb/telegraf/pull/118): Fix for disk usage stats in Windows. Thanks @srfraser!
- [#122](https://github.com/influxdb/telegraf/issues/122): Fix for DiskUsage segv fault. Thanks @srfraser!
- [#126](https://github.com/influxdb/telegraf/issues/126): Nginx plugin not catching net.SplitHostPort error

## v0.1.5 [2015-08-13]

### Features
- [#54](https://github.com/influxdb/telegraf/pull/54): MongoDB plugin. Thanks @jipperinbham!
- [#55](https://github.com/influxdb/telegraf/pull/55): Elasticsearch plugin. Thanks @brocaar!
- [#71](https://github.com/influxdb/telegraf/pull/71): HAProxy plugin. Thanks @kureikain!
- [#72](https://github.com/influxdb/telegraf/pull/72): Adding TokuDB metrics to MySQL. Thanks vadimtk!
- [#73](https://github.com/influxdb/telegraf/pull/73): RabbitMQ plugin. Thanks @ianunruh!
- [#77](https://github.com/influxdb/telegraf/issues/77): Automatically create database.
- [#79](https://github.com/influxdb/telegraf/pull/56): Nginx plugin. Thanks @codeb2cc!
- [#86](https://github.com/influxdb/telegraf/pull/86): Lustre2 plugin. Thanks srfraser!
- [#91](https://github.com/influxdb/telegraf/pull/91): Unit testing
- [#92](https://github.com/influxdb/telegraf/pull/92): Exec plugin. Thanks @alvaromorales!
- [#98](https://github.com/influxdb/telegraf/pull/98): LeoFS plugin. Thanks @mocchira!
- [#103](https://github.com/influxdb/telegraf/pull/103): Filter by metric tags. Thanks @srfraser!
- [#106](https://github.com/influxdb/telegraf/pull/106): Options to filter plugins on startup. Thanks @zepouet!
- [#107](https://github.com/influxdb/telegraf/pull/107): Multiple outputs beyong influxdb. Thanks @jipperinbham!
- [#108](https://github.com/influxdb/telegraf/issues/108): Support setting per-CPU and total-CPU gathering.
- [#111](https://github.com/influxdb/telegraf/pull/111): Report CPU Usage in cpu plugin. Thanks @jpalay!

### Bugfixes
- [#85](https://github.com/influxdb/telegraf/pull/85): Fix GetLocalHost testutil function for mac users
- [#89](https://github.com/influxdb/telegraf/pull/89): go fmt fixes
- [#94](https://github.com/influxdb/telegraf/pull/94): Fix for issue #93, explicitly call sarama.v1 -> sarama
- [#101](https://github.com/influxdb/telegraf/issues/101): switch back from master branch if building locally
- [#99](https://github.com/influxdb/telegraf/issues/99): update integer output to new InfluxDB line protocol format

## v0.1.4 [2015-07-09]

### Features
- [#56](https://github.com/influxdb/telegraf/pull/56): Update README for Kafka plugin. Thanks @EmilS!

### Bugfixes
- [#50](https://github.com/influxdb/telegraf/pull/50): Fix init.sh script to use telegraf directory. Thanks @jseriff!
- [#52](https://github.com/influxdb/telegraf/pull/52): Update CHANGELOG to reference updated directory. Thanks @benfb!

## v0.1.3 [2015-07-05]

### Features
- [#35](https://github.com/influxdb/telegraf/pull/35): Add Kafka plugin. Thanks @EmilS!
- [#47](https://github.com/influxdb/telegraf/pull/47): Add RethinkDB plugin. Thanks @jipperinbham!

### Bugfixes
- [#45](https://github.com/influxdb/telegraf/pull/45): Skip disk tags that don't have a value. Thanks @jhofeditz!
- [#43](https://github.com/influxdb/telegraf/pull/43): Fix bug in MySQL plugin. Thanks @marcosnils!

## v0.1.2 [2015-07-01]

### Features
- [#12](https://github.com/influxdb/telegraf/pull/12): Add Linux/ARM to the list of built binaries. Thanks @voxxit!
- [#14](https://github.com/influxdb/telegraf/pull/14): Clarify the S3 buckets that Telegraf is pushed to.
- [#16](https://github.com/influxdb/telegraf/pull/16): Convert Redis to use URI, support Redis AUTH. Thanks @jipperinbham!
- [#21](https://github.com/influxdb/telegraf/pull/21): Add memcached plugin. Thanks @Yukki!

### Bugfixes
- [#13](https://github.com/influxdb/telegraf/pull/13): Fix the packaging script.
- [#19](https://github.com/influxdb/telegraf/pull/19): Add host name to metric tags. Thanks @sherifzain!
- [#20](https://github.com/influxdb/telegraf/pull/20): Fix race condition with accumulator mutex. Thanks @nkatsaros!
- [#23](https://github.com/influxdb/telegraf/pull/23): Change name of folder for packages. Thanks @colinrymer!
- [#32](https://github.com/influxdb/telegraf/pull/32): Fix spelling of memoory -> memory. Thanks @tylernisonoff!

## v0.1.1 [2015-06-19]

### Release Notes

This is the initial release of Telegraf.
