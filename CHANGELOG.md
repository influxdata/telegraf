## v0.1.5 [unreleased]

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

### Bugfixes
- [#85](https://github.com/influxdb/telegraf/pull/85): Fix GetLocalHost testutil function for mac users
- [#89](https://github.com/influxdb/telegraf/pull/89): go fmt fixes
- [#94](https://github.com/influxdb/telegraf/pull/94): Fix for issue #93, explicitly call sarama.v1 -> sarama
- [#101](https://github.com/influxdb/telegraf/issues/101): switch back from master branch if building locally

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
